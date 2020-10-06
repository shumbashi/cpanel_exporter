package main 

import(
    "net/http"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    "time"
    "flag"
    "path/filepath"
    "log"
    "strconv"
    "strings"

)



var port string


var (
    interval string
    interval_heavy string
    //reg         = prometheus.NewRegistry()
 
   // reg.MustRegister(version.NewCollector("cpanel_exporter"))
  //  if err := r.Register(nc); err != nil {
//		return nil, fmt.Errorf("couldn't register node collector: %s", err)//
//	}
    
    //factory     = promauto.With(reg)
    
    activeUsers = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "cpanel_users_active",
		Help: "Current Active Users",
	})
	
	suspendedUsers = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "cpanel_users_suspended",
		Help: "Current Active Users",
	})
	
    domainsConfigured = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "cpanel_domains_configured",
		Help: "Current Domains and Subdomains setup",
	})
	

	//requestCount.WithLabelValues().Add
	//requestCount.With(prometheus.Labels{"type": "delete", "user": "alice"}).Inc()
	
    cpanelMeta = promauto.NewCounterVec(
        		prometheus.CounterOpts{
        			Name: "cpanel_meta",
        			Help: "cPanel Metadata",
        		},
        		[]string{"version","release"},
    )
    

    
    	
    cpanelPlans = promauto.NewGaugeVec(
        		prometheus.GaugeOpts{
        			Name: "cpanel_plans",
        			Help: "cPanel Plans Configured",
        		},
        		[]string{"plan","owner"},
    )
    
    
    cpanelBandwidth = promauto.NewGaugeVec(
        		prometheus.GaugeOpts{
        			Name: "cpanel_bandwidth",
        			Help: "cPanel Bandwidth Used",
        		},
        		[]string{"user"},
    )
    
    cpanelQuota = promauto.NewGaugeVec(
        		prometheus.GaugeOpts{
        			Name: "cpanel_quota_percent",
        			Help: "cPanel Quota Percent Used",
        		},
        		[]string{"user"},
    )
    
    
    cpanelQuotaUsed = promauto.NewGaugeVec(
        		prometheus.GaugeOpts{
        			Name: "cpanel_quota_used",
        			Help: "cPanel Quota Value Used",
        		},
        		[]string{"user"},
    )
    
    cpanelMailboxes = promauto.NewGauge(
        		prometheus.GaugeOpts{
        			Name: "cpanel_mailboxes_configured",
        			Help: "cPanel Mailboxes",
        		},
    )
    
    cpanelFTP = promauto.NewGauge(
        		prometheus.GaugeOpts{
        			Name: "cpanel_ftp_accounts",
        			Help: "cPanel FTP Accounts",
        		},
    )
    
    cpanelSessionsEmail = promauto.NewGauge(
        		prometheus.GaugeOpts{
        			Name: "cpanel_sessions_email",
        			Help: "cPanel Webmail Session",
        		},
        		
    )
    
     cpanelSessionsWeb = promauto.NewGauge(
        		prometheus.GaugeOpts{
        			Name: "cpanel_sessions_web",
        			Help: "cPanel Admin Sessions",
        		},
        		
    )
)


func fetchMetrics(){
    
    
        dur,err := time.ParseDuration((interval+"s"))
        
        if(err!=nil){
         log.Fatal(err)
        }
        
        for _ = range time.Tick(dur) {
        
         runMetrics()
               
        }
                
    
}

func fetchUapiMetrics() {
    
        dur,err := time.ParseDuration((interval_heavy+"s"))
        
        if(err!=nil){
         log.Fatal(err)
        }
        
        for _ = range time.Tick(dur) {
        
        
         //these are heavier
         runUapiMetrics()
               
        }
}

func runUapiMetrics(){ 
    
    
      for _,u := range getUsernames() {
                        
                       us := filepath.Base(u)
                
                       bw := getBandwidth(us)   
                                           
                       cpanelBandwidth.With(prometheus.Labels{"user": us }).Set(float64(bw))
                       
                       _,used,perc := getQuota(us)
                       
                       fused,_ := strconv.ParseFloat(used,64)
                       
                       cpanelQuota.With(prometheus.Labels{"user": us }).Set(perc)   
                       cpanelQuotaUsed.With(prometheus.Labels{"user": us }).Set(fused) 
      }
    
}

func runMetrics(){
    
               users := getUsers("")
               
               suspended := getUsers("suspended")

               vers := cpanelVersion()

               plans := getPlansWithOwner()

               domains := getDomains()
               
               domains_ct := len(domains)
           
               wsess := getSessions("web")
               
               esess := getSessions("email")
 
               emails := getEmails()
                              
               domainsConfigured.Set(float64(domains_ct))
               
               cpanelFTP.Set(float64(len(getFTP())))
               
               activeUsers.Set(float64(users))
               
               cpanelMailboxes.Set(float64(len(emails)))
               
               suspendedUsers.Set(float64(suspended))
               
               cpanelMeta.With(prometheus.Labels{"version": vers, "release": getRelease() })
               
               cpanelSessionsEmail.Set(float64(esess))
               cpanelSessionsWeb.Set(float64(wsess))
             
               for p,ct := range plans {
                       parts := strings.Split(p,",")                   
                       cpanelPlans.With(prometheus.Labels{"plan": parts[0],"owner": parts[1] }).Set(float64(ct))
               }
               
             
    
}


func main(){
    
        log.SetFlags(log.LstdFlags | log.Lshortfile)
        
        flag.StringVar(&port, "port", "59117", "Metrics Port")
        flag.StringVar(&interval, "interval","60", "Check interval duration 60s by default")
        flag.StringVar(&interval_heavy, "interval_heavy","1800", "Bandwidth and other heavy checks interval, 1800s (30min) by default")
        flag.Parse()
        
        go runMetrics()
        go runUapiMetrics()
        
        go fetchMetrics()
        go fetchUapiMetrics()
    
        http.Handle("/metrics", promhttp.Handler())
        http.ListenAndServe(":"+port, nil)
    
}