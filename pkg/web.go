package pkg

import (
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	"net/http"
	"sort"
	"strings"
	"time"
)

func Web() {
	fmt.Printf("%v - [Starting web]\n", time.Now())
	go getInstances()

	http.HandleFunc("/all/", httpRoute)
	http.HandleFunc("/spot/", httpRoute)
	http.HandleFunc("/spot/wasted/", httpRoute)
	http.HandleFunc("/wasted/", httpRoute)
	http.HandleFunc("/wasted/spot/", httpRoute)

	http.HandleFunc("/region/", httpRoute)
	http.HandleFunc("/region/spot/", httpRoute)
	http.HandleFunc("/region/wasted/", httpRoute)

	http.HandleFunc("/app/", httpRoute)
	http.HandleFunc("/app/spot/", httpRoute)
	http.HandleFunc("/app/wasted/", httpRoute)

	http.HandleFunc("/product/", httpRoute)
	http.HandleFunc("/product/spot/", httpRoute)
	http.HandleFunc("/product/wasted/", httpRoute)

	http.HandleFunc("/env/", httpRoute)
	http.HandleFunc("/env/spot/", httpRoute)
	http.HandleFunc("/env/wasted/", httpRoute)

	http.HandleFunc("/type/", httpRoute)
	http.HandleFunc("/type/spot/", httpRoute)
	http.HandleFunc("/type/wasted/", httpRoute)

	http.HandleFunc("/current/", httpSingleKey)
	http.HandleFunc("/json", httpSingleKey)
	http.HandleFunc("/alerts/", httpSingleKey)
	http.HandleFunc("/alerts/asg/", httpSingleKey)

	http.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "views/urls.html")
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "views/urls.html")
	})

	port := SERVICE_PORT
	fmt.Printf("listening on %v...\n", port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		panic(err)
	}
}

func httpSingleKey(w http.ResponseWriter, r *http.Request) {
	rc := redis.NewClient(&redis.Options{
		Addr:     REDIS_SERVER,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	cutIt := false
	prefix := ""
	switch r.RequestURI {
	case "/current/", "/current/json":
		prefix = "current"
	case "/alerts/", "/alerts/json":
		prefix = "alertas"
	case "/alerts/asg/", "/alerts/asg/json":
		prefix = "alertasasg"
	case "/json":
		prefix = "virginator"
		cutIt = true
	}

	keys, _ := rc.ZRangeWithScores(prefix, 0, -1).Result()

	isJson := false
	if strings.HasSuffix(r.RequestURI, "json") {
		isJson = true
		w.Header()["Access-Control-Allow-Origin"] = []string{"*"}
		w.Header()["Content-Type"] = []string{"application/json"}
	} else {
		fmt.Fprintf(w, "<table>")
	}

	var members []string
	jsonmapcut := make(map[string]map[string]float64)
	mapa := make(map[string]float64)

	for _, k := range keys {
		members = append(members, k.Member.(string))
		mapa[k.Member.(string)] = k.Score

	}

	jsonmap := make(map[string]float64)
	sort.Strings(members)

	for _, k := range members {
		if isJson {
			if cutIt {
				name := strings.SplitN(k, "-", 2)
				a := make(map[string]float64)
				a["us"] = mapa[name[0]+"-us"]
				a["sa"] = mapa[name[0]+"-sa"]
				jsonmapcut[name[0]] = a
			} else {
				jsonmap[k] = mapa[k]
			}
		} else {
			red := ""
			if strings.Contains(k, "none") {
				red = "<font color=red>"
			}

			fmt.Fprintf(w, "<tr><td><big>%v<b>%v :</b></td><td><big>%v</td></tr>", red, k, mapa[k])
		}

	}

	jsonread := ""
	var jsonout []byte
	if isJson {
		if cutIt {
			jsonout, _ = json.Marshal(jsonmapcut)
		} else {
			jsonout, _ = json.Marshal(jsonmap)
		}
		jsonread = string(jsonout[:])
		fmt.Fprintf(w, jsonread)
	} else {
		fmt.Fprintf(w, "</table>")
	}
}

func httpRoute(w http.ResponseWriter, r *http.Request) {
	prefix := "/"
	noit := false
	switch r.RequestURI {
	case "/all/", "/all/json":
		prefix = "r_*"
		noit = true
	case "/spot/", "/spot/json":
		prefix = "s_*"
		noit = true
	case "/wasted/", "/wasted/json":
		prefix = "w_*"
		noit = true
	case "/spot/wasted/", "/sort/wasted/json":
		prefix = "s_wasted*"
	case "/wasted/spot/", "/wasted/spot/json":
		prefix = "w_spot*"
	case "/region/", "/region/json":
		prefix = "r_region*"
	case "/region/spot/", "/region/spot/json":
		prefix = "s_region*"
	case "/region/wasted/", "/region/wasted/json":
		prefix = "w_region*"
	case "/app/", "/app/json":
		prefix = "r_app*"
	case "/app/spot/", "/app/spot/json":
		prefix = "s_app*"
	case "/app/wasted/", "/app/wasted/json":
		prefix = "w_app*"
	case "/product/", "/product/json":
		prefix = "r_product*"
	case "/product/spot/", "/product/spot/json":
		prefix = "s_product*"
	case "/product/wasted/", "/product/wasted/json":
		prefix = "w_product*"
	case "/env/", "/env/json":
		prefix = "r_env*"
	case "/env/spot/", "/env/spot/json":
		prefix = "s_env*"
	case "/env/wasted/", "/env/wasted/json":
		prefix = "w_env*"
	case "/type/", "/type/json":
		prefix = "r_type*" //nice!
	case "/type/spot/", "/type/spot/json":
		prefix = "s_type*"
	case "/type/wasted/", "/type/wasted/json":
		prefix = "w_type*"
	}

	rc := redis.NewClient(&redis.Options{
		Addr:     REDIS_SERVER,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	isJson := false
	if strings.HasSuffix(r.RequestURI, "json") {
		isJson = true
		w.Header()["Access-Control-Allow-Origin"] = []string{"*"}
		w.Header()["Content-Type"] = []string{"application/json"}
	} else {
		fmt.Fprintf(w, "<table>")
	}

	keys, _ := rc.Keys(prefix).Result()
	sort.Strings(keys)
	jsonmap := make(map[string]int64)
	for _, k := range keys {

		value, _ := rc.ZCount(k, "-inf", "+inf").Result()

		name := strings.SplitN(k, "-", 2)

		if isJson {
			if noit {
				jsonmap[k] = value
			} else {
				jsonmap[name[1]] = value
			}

		} else {
			red := ""
			if strings.Contains(k, "none") {
				red = "<font color=red>"
			}
			if noit {
				fmt.Fprintf(w, "<tr><td><big>%v<b>%v :</b></td><td><big>%v</td></tr>", red, k, value)
			} else {
				fmt.Fprintf(w, "<tr><td><big>%v<b>%v :</b></td><td><big>%v</td></tr>", red, name[1], value)
			}
		}
	}

	if isJson {
		jsonout, _ := json.Marshal(jsonmap)
		jsonread := string(jsonout[:])
		fmt.Fprintf(w, jsonread)
	} else {
		fmt.Fprintf(w, "</table>")
	}
}
