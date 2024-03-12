package handle

import (
	"context"
	"gf-sdk-server/src/loger"
	"gf-sdk-server/src/module"
	"github.com/bnb-chain/greenfield-go-sdk/client"
	"github.com/bnb-chain/greenfield-go-sdk/types"
	sdktypes "github.com/bnb-chain/greenfield-go-sdk/types"
	types2 "github.com/bnb-chain/greenfield/x/storage/types"
	"strings"
	"time"
)

func CreateObject(req module.PutObject, bucket, object, id string, cli client.IClient) (string, error) {
	reader := strings.NewReader(req.Data)

	req.Check()

	var v types2.VisibilityType
	switch req.Visibility {
	case 0:
		v = types2.VISIBILITY_TYPE_UNSPECIFIED
	case 1:
		v = types2.VISIBILITY_TYPE_PUBLIC_READ
	case 2:
		v = types2.VISIBILITY_TYPE_PRIVATE
	case 3:
		v = types2.VISIBILITY_TYPE_INHERIT
	}

	cbOpt := types.CreateObjectOptions{
		Visibility:          v,
		TxOpts:              nil,
		SecondarySPAccs:     nil,
		ContentType:         req.ContentType,
		IsReplicaType:       false,
		IsAsyncMode:         false,
		IsSerialComputeMode: false,
		Tags:                nil,
	}

	txnHash, err := cli.CreateObject(context.Background(), bucket, object, reader, cbOpt)
	if err != nil {
		loger.Logger.Errorf("id: [%v], %v", id, err)
		return "", err
	}

	loger.Logger.Infof("id: [%v], create object tx hash: %v", id, txnHash)
	return txnHash, nil
}

func PutObject(req module.PutObject, bucket, object, id string, cli client.IClient) error {
	reader := strings.NewReader(req.Data)

	req.Check()

	objectSize := int64(len(req.Data))
	progressReader := &module.ProgressReader{
		Reader:      reader,
		Total:       objectSize,
		StartTime:   time.Now(),
		LastPrinted: time.Now(),
	}

	putOpt := sdktypes.PutObjectOptions{
		ContentType:      req.ContentType,
		TxnHash:          "",
		DisableResumable: false,
		PartSize:         0,
	}

	err := cli.PutObject(context.Background(), bucket, object, objectSize, progressReader, putOpt)
	if err != nil {
		loger.Logger.Errorf("id: [%v], %v", id, err)
		return err
	}

	return nil
}
