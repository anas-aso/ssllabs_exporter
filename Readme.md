# SSLLabs Client

Getting deep analysis of the configuration of any SSL web server on the public Internet

This client relays the target server hostname to [SSLLabs API](https://www.ssllabs.com/ssltest) and parses the result. It covers retries in case of failures and simplifies the assessment result.

## SSLLabs
> SSL Labs is a non-commercial research effort, run by [Qualys](https://www.qualys.com/), to better understand how SSL, TLS, and PKI technologies are used in practice.

source: https://www.ssllabs.com/about/assessment.html

This project implements SSLLabs API client that would get you the same results as if you use the [web interface](https://www.ssllabs.com/ssltest/).

## Installing

### *go get*

    $ go get -u github.com/diamonwiggins/ssllabs-client/ssllabs

## Example

### Getting all groups

```golang
import (
	"fmt"

  "github.com/diamonwiggins/ssllabs-client/ssllabs"
  "github.com/sirupsen/logrus"
)

func main() {

  	var log = &logrus.Logger{
		Out:       os.Stdout,
		Formatter: &logrus.JSONFormatter{},
		Hooks:     make(logrus.LevelHooks),
		Level:     logrus.InfoLevel,
	  }

    host := "google.com"

    result, err := ssllabs.Analyze(log, host)
    if err != nil {
    	log.WithFields(logrus.Fields{
    		"target": host,
    		"error":  err,
    	}).Error("assessment failed")
    }

    grade := ssllabs.EndpointsLowestGrade(result.Endpoints)
    if grade != "" {
    	log.WithFields(logrus.Fields{
    		"target": host,
    		"grade":  grade,
    	}).Info("grade retrieved")
    } else {
    	log.WithFields(logrus.Fields{
    		"target": host,
    	}).Info("no grade available")
    }
}
```
