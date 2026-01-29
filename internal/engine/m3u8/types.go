package m3u8

type MasterPlaylist struct { //Master Playlist contains details of VariantStreams each of which is a variant stream
	Variants []VariantStream
}

type VariantStream struct { //Variant streams has different resolution, bitrate, frame rate for the same stream and a URI leading us to the MediaPlaylist for the variant we choose
	URI        string
	Bandwidth  int64
	Resolution string
	Codecs     string
	FrameRate  float64
}

type MediaPlaylist struct { //MediaPlaylist contains details of MediaSegments each of which is a media segment
	TargetDuration int
	MediaSequence  int
	Segments       []MediaSegment
	EndList        bool
}

type MediaSegment struct { //MediaSegment is a media segment which is a chunk of the media file
	URI      string
	Duration float64
	Sequence int
	Key      *EncryptionKey
}

type EncryptionKey struct { //EncryptionKey is used for encryption of the media segments
	Method string
	URI    string
	IV     []byte
}
