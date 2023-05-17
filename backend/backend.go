package backend

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pboehm/ddns/shared"
)

type Backend struct {
	config *shared.Config
	lookup *HostLookup
}

func NewBackend(config *shared.Config, lookup *HostLookup) *Backend {
	return &Backend{
		config: config,
		lookup: lookup,
	}
}

type DomainInfo struct {
	Id             int      `json:"id"`
	Zone           string   `json:"zone"`
	Masters        []string `json:"masters"`
	NotifiedSerial int      `json:"notified_serial"`
	Serial         int      `json:"serial"`
	LastCheck      int64    `json:"last_check"`
	Kind           string   `json:"kind"`
}

func NewDomainInfo(config *shared.Config) *DomainInfo {
	return &DomainInfo{
		Id:             1,
		Zone:           fmt.Sprintf("%s.", config.Domain),
		Masters:        []string{},
		NotifiedSerial: 2,
		Serial:         2,
		LastCheck:      time.Now().Unix(),
		Kind:           "native",
	}
}

func (b *Backend) Run() error {
	r := gin.New()
	r.Use(gin.Recovery())

	if b.config.Verbose {
		r.Use(gin.Logger())
	}

	r.GET("/dnsapi/lookup/:qname/:qtype", func(c *gin.Context) {
		request := &Request{
			QName: strings.TrimRight(c.Param("qname"), "."),
			QType: c.Param("qtype"),
		}

		response, err := b.lookup.Lookup(request)
		if err == nil {
			c.JSON(200, gin.H{
				"result": []*Response{response},
			})
		} else {
			if b.config.Verbose {
				log.Printf("Error during lookup: %v", err)
			}

			c.JSON(200, gin.H{
				"result": false,
			})
		}
	})

	r.GET("/dnsapi/getDomainMetadata/:name/:kind", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"result": []string{"0"},
		})
	})

	r.GET("/dnsapi/getAllDomainMetadata/:name", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"result": gin.H{"PRESIGNED": []string{"0"}},
		})
	})

	r.GET("/dnsapi/getAllDomains", func(c *gin.Context) {

		domainInfo := NewDomainInfo(b.config)

		c.JSON(200, gin.H{
			"result": []*DomainInfo{domainInfo},
		})
	})

	return r.Run(b.config.ListenBackend)
}
