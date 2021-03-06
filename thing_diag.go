// Copyright 2020 thinkgos (thinkgo@aliyun.com).  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package aiot

import (
	"encoding/json"

	"github.com/thinkgos/aliyun-iot/infra"
	"github.com/thinkgos/aliyun-iot/uri"
)

// @see https://help.aliyun.com/document_detail/140585.html?spm=a2c4g.11186623.6.715.7227580bd0P6i5

// Wifi wifi status
type Wifi struct {
	Rssi int `json:"rssi"` // 无线信号接收强度
	Snr  int `json:"snr"`  // 无线信号信噪比
	Per  int `json:"per"`  // 数据丢包率
	// 错误信息。仅当设备检测到网络异常后,上报数据包含该参数。
	// 格式:"type,code,count;type,code,count"
	// type:  错误类型
	// code:  错误原因
	// count: 错误数量
	// @see https://help.aliyun.com/document_detail/140585.html?spm=a2c4g.11186623.6.715.36f1791fcf3FJI#table-fvv-k8u-som
	ErrStats string `json:"err_stats"`
}

// P 包含wifi状态和时间戳
type P struct {
	Wifi Wifi  `json:"wifi"`  // 设备的连网方式为Wi-Fi,该参数值由网络状态的四个指标组成
	Time int64 `json:"_time"` // 时间戳可以为空,当为空时,控制台上设备网络状态不展现采集时间,单位ms
}

// DiagParam diag参数域
type DiagParam struct {
	P interface{} `json:"p"`
	// format:数据格式.仅支持simple,表示数据为精简格式。
	// quantity:数量。取值:
	//      single:表示上报单条数据。
	//      batch:表示上报多条数据,仅用于上报历史数据。
	// time:时间.取值:
	//      now: 表示上报当前数据。
	//      history: 表示上报历史数据
	Model string `json:"model"`
}

// DiagRequest 设备主动上报网络状态请求
type DiagRequest struct {
	ID      uint      `json:"id,string"`
	Version string    `json:"version"`
	Params  DiagParam `json:"params"`
}

func (sf *Client) thingDiagPost(pk, dn string, p interface{}, isNow bool) (*Token, error) {
	var model string

	if !sf.hasDiag {
		return nil, ErrNotSupportFeature
	}
	if !sf.IsActive(pk, dn) {
		return nil, ErrNotActive
	}

	if isNow {
		model = "format=simple|quantity=single|time=now"
	} else {
		model = "format=simple|quantity=batch|time=history"
	}

	id := sf.nextRequestID()
	out, err := json.Marshal(&DiagRequest{
		id,
		sf.version,
		DiagParam{
			p,
			model,
		}})
	if err != nil {
		return nil, err
	}

	sf.Log.Debugf("thing.diag.post @%d", id)
	_uri := uri.URI(uri.SysPrefix, uri.ThingDiagPost, pk, dn)
	if err = sf.Publish(_uri, 1, out); err != nil {
		return nil, err
	}
	return sf.putPending(id), nil
}

// ThingDiagPost 设备主动上报当前网络状态
// request:  /sys/{productKey}/{deviceName}/_thing/diag/post
// response: /sys/{productKey}/{deviceName}/_thing/diag/post_reply
func (sf *Client) ThingDiagPost(pk, dn string, p P) (*Token, error) {
	return sf.thingDiagPost(pk, dn, p, true)
}

// ThingDiagHistoryPost 设备主动上报历史网络状态
func (sf *Client) ThingDiagHistoryPost(pk, dn string, ps []P) (*Token, error) {
	if len(ps) == 0 {
		return nil, ErrInvalidParameter
	}
	return sf.thingDiagPost(pk, dn, ps, false)
}

// ProcThingDialPostReply 处理设备主动上报网络状态回复
// request:   /sys/{productKey}/{deviceName}/_thing/diag/post
// response:  /sys/{productKey}/{deviceName}/_thing/diag/post_reply
// subscribe: /sys/{productKey}/{deviceName}/_thing/diag/post_reply
func ProcThingDialPostReply(c *Client, rawURI string, payload []byte) error {
	uris := uri.Spilt(rawURI)
	if len(uris) < 6 {
		return ErrInvalidURI
	}
	rsp := &Response{}
	err := json.Unmarshal(payload, rsp)
	if err != nil {
		return err
	}

	if rsp.Code != infra.CodeSuccess {
		err = infra.NewCodeError(rsp.Code, rsp.Message)
	}

	c.signalPending(Message{rsp.ID, nil, err})
	c.Log.Debugf("thing.diag.post.reply @%d", rsp.ID)
	pk, dn := uris[1], uris[2]
	return c.cb.ThingDialPostReply(c, err, pk, dn)
}
