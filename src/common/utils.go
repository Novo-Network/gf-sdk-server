package common

import "github.com/gofrs/uuid"

func GenUid() string {
	uid, _ := uuid.NewV7()
	id := uid.String()
	return id
}
