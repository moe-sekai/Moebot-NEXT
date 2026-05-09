package gallery

import "time"

type GalleryInfo struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	Name     string `gorm:"uniqueIndex;size:64;not null" json:"name"`
	Aliases  string `gorm:"type:text;default:'[]'" json:"aliases"` // JSON []string
	Mode     string `gorm:"size:16;default:'edit'" json:"mode"`    // edit / view / off
	CoverPID uint   `gorm:"default:0" json:"cover_pid"`
	PicsDir  string `gorm:"size:512" json:"pics_dir"`
}

func (GalleryInfo) TableName() string { return "gallery_galleries" }

type GalleryPic struct {
	PID       uint   `gorm:"primaryKey;autoIncrement" json:"pid"`
	GallName  string `gorm:"index;size:64;not null" json:"gall_name"`
	Path      string `gorm:"size:512" json:"path"`
	Hash1     string `gorm:"size:32" json:"-"`
	Hash2     string `gorm:"size:512" json:"-"`
	ThumbPath string `gorm:"size:512" json:"thumb_path"`
}

func (GalleryPic) TableName() string { return "gallery_pics" }

type GalleryUploadRecord struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    int64     `gorm:"index" json:"user_id"`
	PIDs      string    `gorm:"type:text" json:"pids"` // JSON []uint
	Reverted  bool      `gorm:"default:false" json:"reverted"`
	CreatedAt time.Time `json:"created_at"`
}

func (GalleryUploadRecord) TableName() string { return "gallery_upload_records" }
