package dashboard_graph

import (
	"github.com/gin-gonic/gin"
	cutils "github.com/open-falcon/falcon-plus/common/utils"
	h "github.com/open-falcon/falcon-plus/modules/api/app/helper"
	m "github.com/open-falcon/falcon-plus/modules/api/app/model/dashboard"
	"sort"
	"strings"
	"time"
)

type APITmpGraphCreateReqData struct {
	Endpoints []string `json:"endpoints" binding:"required"`
	Counters  []string `json:"counters" binding:"required"`
}

func DashboardTmpGraphCreate(c *gin.Context) {
	var inputs APITmpGraphCreateReqData
	if err := c.Bind(&inputs); err != nil {
		h.JSONR(c, badstatus, err)
		return
	}

	es := inputs.Endpoints
	cs := inputs.Counters
	sort.Strings(es)
	sort.Strings(cs)

	es_string := strings.Join(es, TMP_GRAPH_FILED_DELIMITER)
	cs_string := strings.Join(cs, TMP_GRAPH_FILED_DELIMITER)
	ck := cutils.Md5(es_string + ":" + cs_string)

	dt := db.Dashboard.Exec("insert ignore into `tmp_graph` (endpoints, counters, ck) values(?, ?, ?) on duplicate key update time_=?", es_string, cs_string, ck, time.Now())
	if dt.Error != nil {
		h.JSONR(c, badstatus, dt.Error)
		return
	}

	tmp_graph := m.DashboardTmpGraph{}
	dt = db.Dashboard.Table("tmp_graph").Where("ck=?", ck).First(&tmp_graph)
	if dt.Error != nil {
		h.JSONR(c, badstatus, dt.Error)
		return
	}

	h.JSONR(c, map[string]int{"id": int(tmp_graph.ID)})
}

func DashboardTmpGraphQuery(c *gin.Context) {
	id := c.Param("id")

	tmp_graph := m.DashboardTmpGraph{}
	dt := db.Dashboard.Table("tmp_graph").Where("id = ?", id).First(&tmp_graph)
	if dt.Error != nil {
		h.JSONR(c, badstatus, dt.Error)
		return
	}

	es := strings.Split(tmp_graph.Endpoints, TMP_GRAPH_FILED_DELIMITER)
	cs := strings.Split(tmp_graph.Counters, TMP_GRAPH_FILED_DELIMITER)

	ret := map[string][]string{
		"endpoints": es,
		"counters":  cs,
	}

	h.JSONR(c, ret)
}

type APIGraphCreateReqData struct {
	ScreenId  int      `json:"screen_id" binding:"required"`
	Title     string   `json:"title" binding:"required"`
	Endpoints []string `json:"endpoints" binding:"required"`
	Counters  []string `json:"counters" binding:"required"`
	TimeSpan  int      `json:"timespan"`
	GraphType string   `json:"graph_type"`
	Method    string   `json:"method"`
	Position  int      `json:"position"`
}

func DashboardGraphCreate(c *gin.Context) {
	var inputs APIGraphCreateReqData
	if err := c.Bind(&inputs); err != nil {
		h.JSONR(c, badstatus, err)
		return
	}

	es := inputs.Endpoints
	cs := inputs.Counters
	sort.Strings(es)
	sort.Strings(cs)
	es_string := strings.Join(es, TMP_GRAPH_FILED_DELIMITER)
	cs_string := strings.Join(cs, TMP_GRAPH_FILED_DELIMITER)

	d := &m.DashboardGraph{
		Title:     inputs.Title,
		Hosts:     es_string,
		Counters:  cs_string,
		ScreenId:  int64(inputs.ScreenId),
		TimeSpan:  inputs.TimeSpan,
		GraphType: inputs.GraphType,
		Method:    inputs.Method,
		Position:  inputs.Position,
	}
	if d.TimeSpan == 0 {
		d.TimeSpan = 3600
	}
	if d.GraphType == "" {
		d.GraphType = "h"
	}

	dt := db.Dashboard.Table("dashboard_graph").Create(&d)
	if dt.Error != nil {
		h.JSONR(c, badstatus, dt.Error)
		return
	}

	h.JSONR(c, "ok")

}

func DashboardGraphGet(c *gin.Context) {
}
