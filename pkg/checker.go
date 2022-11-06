package pkg

import (
	"sync"
)

func CheckDNSTarget(config Config) {
	var wg sync.WaitGroup
	for _, domain := range config.Domains {
		wg.Add(1)
		go func(domain Domain, ses SES) {
			defer wg.Done()
			checkDNSTarget(domain, ses)
		}(domain, config.SES)
	}
	wg.Wait()
}
