package main

type ZapDlAsset struct {
	Name 		string 	`json:"name"`
	Download 	string 	`json:"download"`
	Size 		string 	`size:"size"`
}

type ZapRelease struct {
	Roll			int
	Id 				string		`json:"id"`
	Author 			string		`json:"author"`
	PreRelease 		bool		`json:"prerelease"`
	Releases 		string		`json:"releases"`
	Assets 			ZapDlAsset	`json:"assets"`
	Tag 			string		`json:"tag"`
	PublishedAt 	string		`json:"published_at"`
}

type ZapSource struct {
	Type 		string
	Url			string
}

type ZapReleases struct {
	Releases 	[]ZapRelease
	Author		string
	Source		ZapSource
}