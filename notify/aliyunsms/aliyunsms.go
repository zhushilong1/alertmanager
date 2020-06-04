// Copyright 2019 Prometheus Team
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package aliyunsms

import (
	"context"
	"fmt"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/pkg/errors"

	"github.com/prometheus/alertmanager/config"
	"github.com/prometheus/alertmanager/notify"
	"github.com/prometheus/alertmanager/template"
	"github.com/prometheus/alertmanager/types"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/dysmsapi"
)

// Notifier implements a Notifier for aliyunsms notifications.
type Notifier struct {
	conf   *config.AliyunSmsConfig
	tmpl   *template.Template
	logger log.Logger
	client *dysmsapi.Client

	accessKeyId  string
	accessSecret string
}

type aliyunSmsResponse struct {
	Message   string `json:"Message"`
	RequestId string `json:"RequestId"`
	Code      string `json:"Code"`
}

// New returns a new AliyunSms notifier.
func New(c *config.AliyunSmsConfig, t *template.Template, l log.Logger) (*Notifier, error) {
	client, err := dysmsapi.NewClientWithAccessKey("cn-hangzhou", c.AccessKeyId, c.AccessSecret)
	if err != nil {
		return nil, err
	}

	return &Notifier{conf: c, tmpl: t, logger: l, client: client}, nil
}

// Notify implements the Notifier interface.
func (n *Notifier) Notify(ctx context.Context, as ...*types.Alert) (bool, error) {
	key, err := notify.ExtractGroupKey(ctx)
	if err != nil {
		return false, err
	}

	level.Debug(n.logger).Log("incident", key)
	data := notify.GetTemplateData(ctx, n.tmpl, as, n.logger)

	tmpl := notify.TmplText(n.tmpl, data, &err)
	fmt.Print(data)
	if err != nil {
		return false, err
	}
	// if err != nil {
	// 	return false, fmt.Errorf("templating error: %s", err)
	// }

	request := dysmsapi.CreateSendSmsRequest()
	request.Scheme = "https"

	request.PhoneNumbers = tmpl(n.conf.ToUsers)
	request.SignName = "五二九七"
	request.TemplateCode = "SMS_162733468"

	request.TemplateParam = `{"code":"123455"}`

	resp, err := n.client.SendSms(request)

	if err != nil {
		return false, err
	}

	if resp.Code == "OK" {
		return true, errors.New(resp.Message)
	}

	return false, errors.New(resp.Message)
}
