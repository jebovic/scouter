package imagefetch

import (
	"context"
	"log"
)

// PurgeJob is a robfig/cron-compatible func that enforces the MinIO quota.
// quotaMB is the total allowed storage in mebibytes.
func PurgeJob(ctx context.Context, uploader *Uploader, repo *Repository, quotaMB int64) func() {
	quotaBytes := quotaMB * 1024 * 1024
	return func() {
		objs, err := uploader.ListObjects(ctx)
		if err != nil {
			log.Printf("imagefetch purge: list objects: %v", err)
			return
		}

		var total int64
		for _, o := range objs {
			total += o.Size
		}

		threshold := int64(float64(quotaBytes) * 0.90) // evict until below 90%
		if total <= quotaBytes {
			return // within quota
		}

		log.Printf("imagefetch purge: usage %d MB exceeds quota %d MB — evicting oldest objects",
			total/(1024*1024), quotaMB)

		for _, obj := range objs { // already sorted oldest-first by ListObjects
			if total <= threshold {
				break
			}
			if err := uploader.Delete(ctx, obj.Key); err != nil {
				log.Printf("imagefetch purge: delete object %s: %v", obj.Key, err)
				continue
			}
			if err := repo.DeleteByKey(ctx, obj.Key); err != nil {
				log.Printf("imagefetch purge: delete DB row for %s: %v", obj.Key, err)
			}
			total -= obj.Size
		}
	}
}
