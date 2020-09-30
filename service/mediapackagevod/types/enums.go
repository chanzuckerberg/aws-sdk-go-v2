// Code generated by smithy-go-codegen DO NOT EDIT.

package types

type PeriodTriggersElement string

// Enum values for PeriodTriggersElement
const (
	PeriodTriggersElementAds PeriodTriggersElement = "ADS"
)

type AdMarkers string

// Enum values for AdMarkers
const (
	AdMarkersNone            AdMarkers = "NONE"
	AdMarkersScte35_enhanced AdMarkers = "SCTE35_ENHANCED"
	AdMarkersPassthrough     AdMarkers = "PASSTHROUGH"
)

type EncryptionMethod string

// Enum values for EncryptionMethod
const (
	EncryptionMethodAes_128    EncryptionMethod = "AES_128"
	EncryptionMethodSample_aes EncryptionMethod = "SAMPLE_AES"
)

type ManifestLayout string

// Enum values for ManifestLayout
const (
	ManifestLayoutFull    ManifestLayout = "FULL"
	ManifestLayoutCompact ManifestLayout = "COMPACT"
)

type Profile string

// Enum values for Profile
const (
	ProfileNone      Profile = "NONE"
	ProfileHbbtv_1_5 Profile = "HBBTV_1_5"
)

type SegmentTemplateFormat string

// Enum values for SegmentTemplateFormat
const (
	SegmentTemplateFormatNumber_with_timeline SegmentTemplateFormat = "NUMBER_WITH_TIMELINE"
	SegmentTemplateFormatTime_with_timeline   SegmentTemplateFormat = "TIME_WITH_TIMELINE"
	SegmentTemplateFormatNumber_with_duration SegmentTemplateFormat = "NUMBER_WITH_DURATION"
)

type StreamOrder string

// Enum values for StreamOrder
const (
	StreamOrderOriginal                 StreamOrder = "ORIGINAL"
	StreamOrderVideo_bitrate_ascending  StreamOrder = "VIDEO_BITRATE_ASCENDING"
	StreamOrderVideo_bitrate_descending StreamOrder = "VIDEO_BITRATE_DESCENDING"
)