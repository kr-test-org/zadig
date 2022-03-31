package service

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"

	commonrepo "github.com/koderover/zadig/pkg/microservice/aslan/core/common/repository/mongodb"
	"github.com/koderover/zadig/pkg/setting"
	e "github.com/koderover/zadig/pkg/tool/errors"
	toolssh "github.com/koderover/zadig/pkg/tool/ssh"
	"github.com/koderover/zadig/pkg/tool/wsconn"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func ConnectSshPmExec(c *gin.Context, username, envName, productName, ip string, cols, rows int, log *zap.SugaredLogger) error {
	resp, err := commonrepo.NewPrivateKeyColl().Find(commonrepo.FindPrivateKeyOption{
		Address: ip,
	})
	if err != nil {
		log.Errorf("PrivateKey.Find ip %s error: %s", ip, err)
		return e.ErrGetPrivateKey
	}
	if resp.Status != setting.PMHostStatusNormal {
		return e.ErrLoginPm.AddDesc(fmt.Sprintf("host %s status %s,is not normal", ip, resp.Status))
	}
	if resp.Port == 0 {
		resp.Port = setting.PMHostDefaultPort
	}

	sDec, err := base64.StdEncoding.DecodeString(resp.PrivateKey)
	if err != nil {
		log.Errorf("base64 decode failed ip:%s, error:%s", ip, err)
		return e.ErrLoginPm.AddDesc(fmt.Sprintf("base64 decode failed ip:%s, error:%s", ip, err))
	}

	sshCli, err := toolssh.NewSshCli(sDec, resp.UserName, resp.IP, resp.Port)
	if err != nil {
		log.Errorf("NewSshCli err:%s", err)
		return e.ErrLoginPm.AddErr(err)
	}

	sshConn, err := wsconn.NewSshConn(cols, rows, sshCli)
	if err != nil {
		log.Errorf("NewSshConn err:%s", err)
		return e.ErrLoginPm.AddErr(err)
	}
	defer sshConn.Close()

	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Errorf("ws upgrade err:%s", err)
		return e.ErrLoginPm.AddErr(err)
	}
	defer ws.Close()

	stopChan := make(chan bool, 3)
	var logBuff = new(bytes.Buffer)

	go sshConn.ReadWsMessage(ws, logBuff, stopChan)
	go sshConn.SendWsWriteMessage(ws, stopChan)
	go sshConn.SessionWait(stopChan)

	<-stopChan
	return nil
}
