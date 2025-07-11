package googlecloud_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/smart-unicom/oss/googlecloud"
)

func getClient() *googlecloud.Client {
	serviceAccountJson := `{
  "type": "service_account",
  "project_id": "casbin",
  "private_key_id": "xxx",
  "private_key": "-----BEGIN PRIVATE KEY-----\nxxx\n-----END PRIVATE KEY-----\n",
  "client_email": "casdoor-service-account@casbin.iam.gserviceaccount.com",
  "client_id": "10336152244758146xxx",
  "auth_uri": "https://accounts.google.com/o/oauth2/auth",
  "token_uri": "https://oauth2.googleapis.com/token",
  "auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
  "client_x509_cert_url": "https://www.googleapis.com/robot/v1/metadata/x509/casdoor-service-account%40casbin.iam.gserviceaccount.com",
  "universe_domain": "googleapis.com"
}`

	config := &googlecloud.Config{
		ServiceAccountJson: serviceAccountJson,
		Bucket:             "/smart-unicom",
		Endpoint:           "",
	}

	client, err := googlecloud.New(config)
	if err != nil {
		panic(err)
	}

	return client
}

func TestClientPut(t *testing.T) {
	f, err := ioutil.ReadFile("E:/123.txt")
	if err != nil {
		panic(err)
	}

	client := getClient()
	_, err = client.Put("123.txt", bytes.NewReader(f))
	if err != nil {
		panic(err)
	}
}

func TestClientDelete(t *testing.T) {
	client := getClient()
	err := client.Delete("123.txt")
	if err != nil {
		panic(err)
	}
}

func TestClientList(t *testing.T) {
	client := getClient()
	objects, err := client.List("/")
	if err != nil {
		panic(err)
	}

	fmt.Println(objects)
}

func TestClientGet(t *testing.T) {
	client := getClient()
	f, err := client.Get("/")
	if err != nil {
		panic(err)
	}

	fmt.Println(f)
}
