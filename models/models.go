package models

type Couple struct {
	AnchorTimeMs uint32
	SongID       uint32
}

type RecordData struct {
	Audio      string  `json:"audio"`
	Duration   float64 `json:"duration"`
	Channels   int     `json:"channels"`
	SampleRate int     `json:"sampleRate"`
	SampleSize int     `json:"sampleSize"`
}
