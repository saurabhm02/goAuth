package worker

import (
	"goAuth/internals/types"
	"log"
	"time"
)

func StartOTPCleanup(store map[string]interface{}) {
	ticker := time.NewTicker(2 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		log.Println("[cleanup] starting expired OTP cleanup task")
		for id, pcObj := range store {
			pc, ok := pcObj.(*types.ProjectContext)
			if !ok || !pc.OTP || pc.OTPRepo == nil {
				continue
			}

			if otpRepo, ok := pc.OTPRepo.(types.OTPRepository); ok {
				if err := otpRepo.DeleteExpired(); err != nil {
					log.Printf("[cleanup] error cleaning expired OTPs for project %s: %v", id, err)
				} else {
					log.Printf("[cleanup] expired OTPs cleaned for project %s", id)
				}
			}
		}
	}
}
