package main

import (
	"context"
	"flag"
	"gf-sdk-server/src/common"
	"gf-sdk-server/src/handle"
	"gf-sdk-server/src/loger"
	"gf-sdk-server/src/module"
	"github.com/bnb-chain/greenfield-go-sdk/client"
	"github.com/bnb-chain/greenfield-go-sdk/types"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

func main() {

	// other components init
	{
		loger.InitLogger()
	}

	// get config
	var conf module.Config
	{
		var (
			privateKeyPath string
			host           string
			chainRpc       string
			chainId        string
		)

		flag.StringVar(&privateKeyPath, "private_key_path", "gf-sdk-server.pk", "gf sdk server private key path")
		flag.StringVar(&host, "host", "0.0.0.0:8099", "ip and port")
		flag.StringVar(&chainRpc, "chain_rpc", "https://gnfd-testnet-fullnode-tendermint-us.bnbchain.org:443", "default: testnet rpc")
		flag.StringVar(&chainId, "chain_id", "greenfield_5600-1", "default: testnet chain id")

		flag.Parse()

		conf = module.Config{
			Server: module.ServerCfg{
				PrivateKeyPath: privateKeyPath,
				Host:           host,
				ChainRpc:       chainRpc,
				ChainId:        chainId,
			},
		}
		loger.Logger.Infof("config init: %v", conf)
	}

	// get privateKey
	var privateKey string
	{

		file, err := os.Open(conf.Server.PrivateKeyPath)
		if err != nil {
			log.Fatalf("failed to open file: %s", err)
		}
		defer file.Close()

		bytes, err := ioutil.ReadAll(file)
		if err != nil {
			log.Fatalf("failed to read file: %s", err)
		}
		println(bytes)

		privateKey = strings.TrimSpace(string(bytes))
	}

	// get cli
	var cli client.IClient
	{
		account, err := types.NewAccountFromPrivateKey("gnfd-account", privateKey)
		if err != nil {
			panic(err)
		}

		cli, err = client.New(conf.Server.ChainId, conf.Server.ChainRpc, client.Option{DefaultAccount: account, ForceToUseSpecifiedSpEndpointForDownloadOnly: ""})
		if err != nil {
			panic(err)
		}
	}

	// start server
	{
		r := gin.Default()
		r.Use(cors.Default())

		r.GET("/object/:bucket/:object", func(c *gin.Context) {
			id := common.GenUid()

			bucket := c.Param("bucket")
			object := c.Param("object")
			loger.Logger.Infof("id: [%v], get obj, bucket: [%v], object: [%v]", id, bucket, object)

			reader, contentType, err := cli.GetObject(context.Background(), bucket, object, types.GetObjectOptions{})
			if err != nil {
				loger.Logger.Errorln(err)
				c.JSON(http.StatusInternalServerError, module.NewErrResp(err, id))
				return
			}
			defer reader.Close()

			contentLength := int64(-1)
			c.DataFromReader(http.StatusOK, contentLength, contentType.ContentType, reader, nil)
		})

		r.PUT("/object/:bucket/:object", func(c *gin.Context) {
			id := common.GenUid()

			bucket := c.Param("bucket")
			object := c.Param("object")
			loger.Logger.Infof("id: [%v], put obj, bucket: [%v], object: [%v]", id, bucket, object)

			var req module.PutObject

			// 绑定 JSON 数据
			if err := c.BindJSON(&req); err != nil {
				// 如果解析出错，返回错误信息
				c.JSON(http.StatusInternalServerError, module.NewErrResp(err, id))
				return
			}

			err := handle.PutObject(req, bucket, object, id, cli)
			if err != nil {
				c.JSON(http.StatusInternalServerError, module.NewErrResp(err, id))
				return
			}

			if req.Sync {
				// Check if object is sealed
				timeout := time.After(1 * time.Minute)
				ticker := time.NewTicker(3 * time.Second)
				count := 0
				loger.Logger.Infof("id: [%v], sealing...", id)
				for {
					select {
					case <-timeout:
						loger.Logger.Errorf("id: [%v], %v", id, err)
						c.JSON(http.StatusInternalServerError, module.NewResp("object not sealed after one min", -1, id))
						return
					case <-ticker.C:
						count++
						headObjOutput, queryErr := cli.HeadObject(c, bucket, object)
						if queryErr != nil {
							loger.Logger.Errorf("id: [%v], %v", id, err)
							c.JSON(http.StatusInternalServerError, module.NewErrResp(err, id))
							return
						}
						if count%10 == 0 {
							loger.Logger.Infof("id:[%v], sealing...", id)
						}
						if headObjOutput.ObjectInfo.GetObjectStatus().String() == "OBJECT_STATUS_SEALED" {
							ticker.Stop()
							loger.Logger.Infof("id: [%v], upload %v to %v ", id, object, bucket)
							c.JSON(http.StatusOK, module.Resp{})
							return
						}
					}
				}
			} else {
				c.JSON(http.StatusOK, module.Resp{})
				return
			}
		})

		r.POST("/object/:bucket/:object", func(c *gin.Context) {
			id := common.GenUid()

			bucket := c.Param("bucket")
			object := c.Param("object")
			loger.Logger.Infof("id: [%v], create obj, bucket: [%v], object: [%v]", id, bucket, object)

			var req module.PutObject

			// 绑定 JSON 数据
			if err := c.BindJSON(&req); err != nil {
				// 如果解析出错，返回错误信息
				loger.Logger.Errorf("id: [%v], %v", id, err)
				c.JSON(http.StatusInternalServerError, module.NewErrResp(err, id))
				return
			}

			txnHash, err := handle.CreateObject(req, bucket, object, id, cli)
			if err != nil {
				loger.Logger.Errorf("id: [%v], %v", id, err)
				c.JSON(http.StatusInternalServerError, module.NewErrResp(err, id))
				return
			}

			if req.Sync {
				tx, err := cli.WaitForTx(context.Background(), txnHash)
				if err != nil {
					loger.Logger.Errorf("id: [%v], %v", id, err)
					c.JSON(http.StatusInternalServerError, module.NewErrResp(err, id))
					return
				}
				if tx.TxResult.Code == 0 {
					c.JSON(http.StatusOK, module.Resp{})
					return
				}
			} else {
				c.JSON(http.StatusOK, module.Resp{})
				return
			}

		})

		r.Run(conf.Server.Host)
	}

}
