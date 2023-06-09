package nsq

const (
	TopicMetadata        = "packagist.metadata"
	TopicProvider        = "packagist.provider"
	TopicProviderPackage = "packagist.provider-package"
	CHANNEL              = "packagist-mirrors"
	//TopicPackagistWait   = []string{
	//	TopicProvider,
	//	TopicProviderPackage,
	//}
)

var TopicPackagistWait = []string{
	TopicProvider,
	TopicProviderPackage,
}
