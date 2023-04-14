package filter

import (
	"github.com/Qianlitp/crawlergo/pkg/logger"
	"github.com/Qianlitp/crawlergo/pkg/tools"
	"github.com/Qianlitp/crawlergo/pkg/tools/requests"
	"strings"

	"github.com/Qianlitp/crawlergo/pkg/config"
	"github.com/Qianlitp/crawlergo/pkg/model"
	mapset "github.com/deckarep/golang-set"
)

type SimpleFilter struct {
	UniqueSet mapset.Set
	HostLimit []string
}

var (
	staticSuffixSet = config.StaticSuffixSet.Clone()
)

func init() {
	for _, suffix := range []string{"js", "css", "json"} {
		staticSuffixSet.Add(suffix)
	}
}

/*
*
需要过滤则返回 true
*/
func (s *SimpleFilter) DoFilter(req *model.Request) bool {
	if s.UniqueSet == nil {
		s.UniqueSet = mapset.NewSet()
	}
	// 首先判断是否需要过滤域名

	if len(s.HostLimit) != 0 && s.DomainFilter(req) {
		return true
	}
	// 去重
	if s.UniqueFilter(req) {
		return true
	}
	if s.RequestFilter(req) {
		return true
	}
	// 过滤静态资源
	if s.StaticFilter(req) {
		return true
	}
	return false
}

//过滤不符合条件的请求

func (s *SimpleFilter) RequestFilter(req *model.Request) bool {
	defer func() bool {
		if r := recover(); r != nil {
			logger.Logger.Error("There was an error:", r)
		}
		return true
	}()
	res, err := requests.Get(req.URL.String(), tools.ConvertHeaders(req.Headers),
		&requests.ReqOptions{AllowRedirect: false,
			Timeout: 5,
			Proxy:   req.Proxy})
	// 用于处理报错情况,比如协议是http但是用了https
	if err != nil {
		panic(err)
	}
	// 如果返回404\502过滤该请求
	if res.StatusCode == 404 || res.StatusCode == 502 || res.StatusCode == 407 || res.StatusCode == 503 {
		return true
	}
	return false
}

/*
*
请求去重
*/
func (s *SimpleFilter) UniqueFilter(req *model.Request) bool {
	if s.UniqueSet == nil {
		s.UniqueSet = mapset.NewSet()
	}
	if s.UniqueSet.Contains(req.UniqueId()) {
		return true
	} else {
		s.UniqueSet.Add(req.UniqueId())
		return false
	}
}

/*
*
静态资源过滤
*/
func (s *SimpleFilter) StaticFilter(req *model.Request) bool {
	if s.UniqueSet == nil {
		s.UniqueSet = mapset.NewSet()
	}
	// 首先将slice转换成map

	if req.URL.FileExt() == "" {
		return false
	}
	if staticSuffixSet.Contains(req.URL.FileExt()) {
		return true
	}
	return false
}

/*
*
只保留指定域名的链接
*/
func (s *SimpleFilter) DomainFilter(req *model.Request) bool {
	if s.UniqueSet == nil {
		s.UniqueSet = mapset.NewSet()
	}
	for _, host := range s.HostLimit {
		if req.URL.Host == host || req.URL.Hostname() == host {
			return false
		}
		if strings.HasSuffix(host, ":80") && req.URL.Port() == "" && req.URL.Scheme == "http" {
			if req.URL.Hostname()+":80" == host {
				return false
			}
		}
		if strings.HasSuffix(host, ":443") && req.URL.Port() == "" && req.URL.Scheme == "https" {
			if req.URL.Hostname()+":443" == host {
				return false
			}
		}
	}

	return true
}
