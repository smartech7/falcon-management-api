package graph

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/emirpasic/gods/maps/hashmap"
	"github.com/gin-gonic/gin"
	cmodel "github.com/open-falcon/falcon-plus/common/model"
	h "github.com/open-falcon/falcon-plus/modules/api/app/helper"
	m "github.com/open-falcon/falcon-plus/modules/api/app/model/graph"
	u "github.com/open-falcon/falcon-plus/modules/api/app/utils"
	"regexp"
	"strings"
)

type APIGrafanaMainQueryInputs struct {
	Limit int    `json:"limit"  form:"limit"`
	Query string `json:"query"  form:"query"`
}

type APIGrafanaMainQueryOutputs struct {
	Expandable bool   `json:"expandable"`
	Text       string `json:"text"`
}

//for return a host list for api test
func repsonseDefault(limit int) (result []APIGrafanaMainQueryOutputs) {
	result = []APIGrafanaMainQueryOutputs{}
	//for get right table name
	enpsHelp := m.Endpoint{}
	enps := []m.Endpoint{}
	db.Graph.Table(enpsHelp.TableName()).Limit(limit).Scan(&enps)
	for _, h := range enps {
		result = append(result, APIGrafanaMainQueryOutputs{
			Expandable: true,
			Text:       h.Endpoint,
		})
	}
	return
}

//for find host list & grafana template searching, regexp support
func responseHostsRegexp(limit int, regexpKey string) (result []APIGrafanaMainQueryOutputs) {
	result = []APIGrafanaMainQueryOutputs{}
	//for get right table name
	enpsHelp := m.Endpoint{}
	enps := []m.Endpoint{}
	db.Graph.Table(enpsHelp.TableName()).Where("endpoint regexp ?", regexpKey).Limit(limit).Scan(&enps)
	for _, h := range enps {
		result = append(result, APIGrafanaMainQueryOutputs{
			Expandable: true,
			Text:       h.Endpoint,
		})
	}
	return
}

//for resolve mixed query with endpoint & counter of query string
func cutEndpointCounterHelp(regexpKey string) (hosts []string, counter string) {
	r, _ := regexp.Compile("^{?([^#}]+)}?#(.+)")
	matchedList := r.FindAllStringSubmatch(regexpKey, 1)
	if len(matchedList) != 0 {
		if len(matchedList[0]) > 1 {
			//get hosts
			hostsTmp := matchedList[0][1]
			counterTmp := matchedList[0][2]
			hosts = strings.Split(hostsTmp, ",")
			counter = strings.Replace(counterTmp, "#", "\\.", -1)
		}
	} else {
		log.Errorf("grafana query inputs error: %v", regexpKey)
	}
	return
}

func expandableChecking(counter string, counterSearchKeyWord string) (expsub string, needexp bool) {
	re := regexp.MustCompile("(\\.\\+|\\.\\*)\\s*$")
	counterSearchKeyWord = re.ReplaceAllString(counterSearchKeyWord, "")
	counterSearchKeyWord = strings.Replace(counterSearchKeyWord, "\\.", ".", -1)
	expCheck := strings.Replace(counter, counterSearchKeyWord, "", -1)
	if expCheck == "" {
		needexp = false
		expsub = expCheck
	} else {
		needexp = true
		re = regexp.MustCompile("^\\.")
		expsubArr := strings.Split(re.ReplaceAllString(expCheck, ""), ".")
		switch len(expsubArr) {
		case 0:
			expsub = ""
		case 1:
			expsub = expsubArr[0]
			needexp = false
		default:
			expsub = expsubArr[0]
			needexp = true
		}
	}
	return
}

/* add additional items (ex. $ & %)
   $ means metric is stop on here. no need expand any more.
   % means a wirecard string.
   also clean defecate metrics
*/
func addAddItionalItems(items []APIGrafanaMainQueryOutputs, regexpKey string) (result []APIGrafanaMainQueryOutputs) {
	flag := false
	mapset := hashmap.New()
	for _, i := range items {
		if !i.Expandable {
			flag = true
		}
		if val, exist := mapset.Get(i.Text); exist {
			if val != i.Expandable && i.Expandable {
				mapset.Put(i.Text, i.Expandable)
			}
		} else {
			mapset.Put(i.Text, i.Expandable)
		}
	}
	result = make([]APIGrafanaMainQueryOutputs, mapset.Size())
	for idx, ctmp := range mapset.Keys() {
		c := ctmp.(string)
		val, _ := mapset.Get(c)
		result[idx] = APIGrafanaMainQueryOutputs{
			Text:       c,
			Expandable: val.(bool),
		}
	}
	if flag {
		result = append(result, APIGrafanaMainQueryOutputs{
			Text:       "$",
			Expandable: false,
		})
	}
	if len(strings.Split(regexpKey, "\\.")) > 0 {
		result = append(result, APIGrafanaMainQueryOutputs{
			Text:       "%",
			Expandable: false,
		})
	}
	return
}

func findEndpointIdByEndpointList(hosts []string) []int64 {
	//for get right table name
	enpsHelp := m.Endpoint{}
	enps := []m.Endpoint{}
	hostStr, _ := u.ArrStringsToString(hosts)
	log.Debugf("hostStr: %v", hostStr)
	log.Debugf("endpoint in (%s)", hostStr)
	db.Graph.Table(enpsHelp.TableName()).Where(fmt.Sprintf("endpoint in (%s)", hostStr)).Scan(&enps)
	hostIds := make([]int64, len(enps))
	for indx, h := range enps {
		hostIds[indx] = int64(h.ID)
	}
	return hostIds
}

//for reture counter list of endpoints
func responseCounterRegexp(regexpKey string) (result []APIGrafanaMainQueryOutputs) {
	result = []APIGrafanaMainQueryOutputs{}
	hosts, counter := cutEndpointCounterHelp(regexpKey)
	if len(hosts) == 0 || counter == "" {
		return
	}
	hostIds := findEndpointIdByEndpointList(hosts)
	//if not any endpoint matched
	if len(hostIds) == 0 {
		return
	}
	idConcact, _ := u.ArrInt64ToString(hostIds)
	//for get right table name
	countHelp := m.EndpointCounter{}
	counters := []m.EndpointCounter{}
	db.Graph.Table(countHelp.TableName()).Where(fmt.Sprintf("endpoint_id IN (%s) AND counter regexp '%s'", idConcact, counter)).Scan(&counters)
	//if not any counter matched
	if len(counters) == 0 {
		return
	}
	for _, c := range counters {
		expsub, needexp := expandableChecking(c.Counter, counter)
		result = append(result, APIGrafanaMainQueryOutputs{
			Text:       expsub,
			Expandable: needexp,
		})
	}
	result = addAddItionalItems(result, regexpKey)
	return
}

func GrafanaMainQuery(c *gin.Context) {
	inputs := APIGrafanaMainQueryInputs{}
	inputs.Limit = 1000
	inputs.Query = "!N!"
	if err := c.Bind(&inputs); err != nil {
		h.JSONR(c, badstatus, err.Error())
		return
	}
	log.Debugf("got query string: %s", inputs.Query)
	output := []APIGrafanaMainQueryOutputs{}
	if inputs.Query == "!N!" {
		output = repsonseDefault(inputs.Limit)
	} else if !strings.Contains(inputs.Query, "#") {
		output = responseHostsRegexp(inputs.Limit, inputs.Query)
	} else if strings.Contains(inputs.Query, "#") && !strings.Contains(inputs.Query, "#select metric") {
		output = responseCounterRegexp(inputs.Query)
	}
	h.JSONR(c, output)
	return
}

type APIGrafanaRenderInput struct {
	Target        string `json:"target" form:"target"  binding:"required"`
	From          int64  `json:"from" form:"from" binding:"required"`
	Until         int64  `json:"until" form:"until" binding:"required"`
	Format        string `json:"format" form:"format"`
	MaxDataPoints int64  `json:"maxDataPoints" form:"maxDataPoints"`
	Step          int    `json:"step" form:"step"`
	ConsolFun     string `json:"consolFun" form:"consolFun"`
}

func GrafanaRender(c *gin.Context) {
	inputs := APIGrafanaRenderInput{}
	//set default step is 60
	inputs.Step = 60
	inputs.ConsolFun = "AVERAGE"
	if err := c.Bind(&inputs); err != nil {
		h.JSONR(c, badstatus, err.Error())
		return
	}
	hosts, counter := cutEndpointCounterHelp(inputs.Target)
	//clean characters
	log.Println(counter)
	re := regexp.MustCompile("\\\\.\\$\\s*$")
	flag := re.MatchString(counter)
	counter = re.ReplaceAllString(counter, "")
	counter = strings.Replace(counter, "\\.%", ".+", -1)
	ecHelp := m.EndpointCounter{}
	counters := []m.EndpointCounter{}
	log.Println(counter)
	hostIds := findEndpointIdByEndpointList(hosts)
	if flag {
		db.Graph.Table(ecHelp.TableName()).Select("distinct counter").Where(fmt.Sprintf("endpoint_id IN (%s) AND counter = '%s'", u.ArrInt64ToStringMust(hostIds), counter)).Scan(&counters)
	} else {
		db.Graph.Table(ecHelp.TableName()).Select("distinct counter").Where(fmt.Sprintf("endpoint_id IN (%s) AND counter regexp '%s'", u.ArrInt64ToStringMust(hostIds), counter)).Scan(&counters)
	}
	if len(counters) == 0 {
		h.JSONR(c, []interface{}{})
		return
	}
	counterArr := make([]string, len(counters))
	for indx, c := range counters {
		counterArr[indx] = c.Counter
	}
	respList := []*cmodel.GraphQueryResponse{}
	for _, host := range hosts {
		for _, c := range counterArr {
			resp, err := fetchData(host, c, inputs.ConsolFun, inputs.From, inputs.Until, inputs.Step)
			if err != nil {
				log.Debugf("query graph got error with: %v", inputs)
			} else {
				respList = append(respList, resp)
			}
		}
	}
	h.JSONR(c, respList)
	return
}
