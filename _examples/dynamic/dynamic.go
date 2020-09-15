package main

import (
	"log"

	"github.com/thinkgos/aliyun-iot/_examples/mock"
	"github.com/thinkgos/aliyun-iot/dynamic"
	"github.com/thinkgos/aliyun-iot/infra"
)

func main() {
	meta := &infra.MetaTetrad{
		ProductKey:    mock.ProductKey,
		ProductSecret: mock.ProductSecret,
		DeviceName:    mock.DeviceName,
	}
	crd := infra.CloudRegionDomain{
		Region:       infra.CloudRegionShangHai,
		CustomDomain: "127.0.0.1:8080",
	}

	cli := dynamic.New()

	err := cli.RegisterCloud(meta, crd)
	if err != nil {
		panic(err)
	}
	log.Printf("%+v", meta)
}
