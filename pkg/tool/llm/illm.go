/*
Copyright 2023 The K8sGPT Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Some parts of this file have been modified to make it functional in Zadig

package llm

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/koderover/zadig/v2/pkg/tool/cache"
)

var (
	clients = map[string]ILLM{
		"openai":      &OpenAIClient{},
		"azureopenai": &OpenAIClient{},
	}
)

type ILLM interface {
	Configure(config LLMConfig) error
	GetCompletion(ctx context.Context, prompt string, options ...ParamOption) (string, error)
	Parse(ctx context.Context, prompt string, cache cache.ICache, options ...ParamOption) (string, error)
	GetName() string
}

func NewClient(provider string) (ILLM, error) {
	if c, ok := clients[provider]; !ok {
		return nil, fmt.Errorf("provider %s not supported", provider)
	} else {
		return c, nil
	}
}

type LLMConfig struct {
	Name    string
	Model   string
	Token   string
	BaseURL string
	Proxy   string
	APIType string
}

func (p *LLMConfig) GetName() string {
	return p.Name
}

func (p *LLMConfig) GetBaseURL() string {
	return p.BaseURL
}

func (p *LLMConfig) GetToken() string {
	return p.Token
}

func (p *LLMConfig) GetModel() string {
	return p.Model
}

func (p *LLMConfig) GetProxy() string {
	return p.Proxy
}

func (p *LLMConfig) GetAPIType() string {
	return p.APIType
}

func GetCacheKey(provider string, sEnc string) string {
	data := fmt.Sprintf("%s-%s", provider, sEnc)

	hash := sha256.Sum256([]byte(data))

	return hex.EncodeToString(hash[:])
}
