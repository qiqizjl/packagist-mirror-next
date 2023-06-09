package store

const (
	PackagistLastModified     = "packagist:last_modified"
	PackagistDistVersion      = "packagist:dist:version:%s"
	PackagistMetadata         = "packagist:metadata"
	PackagistProvider         = "packagist:provider"
	PackagistProviderPackage  = "packagist:provider-package"
	PackagistError            = "%s:error:%d"
	PackagistStat             = "%s:stat:%s"
	PackagistMetadataLastSync = "packagist:metadata:last_sync"
	PackagistQueueInfo        = "packagist:queue:info:%s"
	PackagistPackagesLastSync = "packagist:packages:last_sync"
)

// 错误列表
var errorList = []int{
	400,
	401,
	402,
	403,
	404,
	500,
	502,
	504,
}
