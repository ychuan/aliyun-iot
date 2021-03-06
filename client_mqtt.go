// Copyright 2020 thinkgos (thinkgo@aliyun.com).  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package aiot

import (
	"log"

	mqtt "github.com/eclipse/paho.mqtt.golang"

	"github.com/thinkgos/aliyun-iot/infra"
)

// MQTTClient MQTT客户端
type MQTTClient struct {
	c mqtt.Client
	*Client
}

// 确保 NopCb 实现 dm.Conn 接口
var _ Conn = (*MQTTClient)(nil)

// NewWithMQTT 新建MQTTClient
func NewWithMQTT(meta infra.MetaTriad, c mqtt.Client, opts ...Option) *MQTTClient {
	m := New(meta, nil, opts...)
	cli := &MQTTClient{
		c,
		m,
	}
	m.Conn = cli
	return cli
}

// Underlying 获得底层的Client
func (sf *MQTTClient) Underlying() mqtt.Client { return sf.c }

// Publish 实现dm.Conn接口
func (sf *MQTTClient) Publish(topic string, qos byte, payload interface{}) error {
	return sf.c.Publish(topic, qos, false, payload).Error()
}

// Subscribe 实现dm.Conn接口
func (sf *MQTTClient) Subscribe(topic string, streamFunc ProcDownStream) error {
	return sf.c.Subscribe(topic, 1, func(client mqtt.Client, message mqtt.Message) {
		if message.Duplicate() {
			return
		}
		if err := streamFunc(sf.Client, message.Topic(), message.Payload()); err != nil {
			log.Printf("topic: %s, error: %+v\r\n", message.Topic(), err)
		}
	}).Error()
}

// UnSubscribe 实现dm.Conn接口
func (sf *MQTTClient) UnSubscribe(topic ...string) error {
	return sf.c.Unsubscribe(topic...).Error()
}

// Close 实现dm.Conn接口
func (sf *MQTTClient) Close() error {
	sf.c.Disconnect(500)
	return nil
}
