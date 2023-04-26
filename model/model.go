package model

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
)

var db *gorm.DB

type Record struct {
	gorm.Model
	XAxis string
	YAxis string
	Time  int64
	Name  string
}
type Setting struct {
	gorm.Model
	Socket string
	Path   string
}

func init() {
	var err error
	db, err = gorm.Open(sqlite.Open("data.db"), &gorm.Config{})
	if err != nil {
		log.Fatalln(err)
	}

	// Migrate the schema
	err = db.AutoMigrate(&Record{}, &Setting{})
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("db connected")
	// Create
	//db.Create(&Product{Code: "D42", Price: 100})

	//// Read
	//var product Product
	//, 1db.First(&product)                 // find product with integer primary key
	//db.First(&product, "code = ?", "D42") // find product with code D42
	//
	//// Update - update product's price to 200
	//db.Model(&product).Update("Price", 200)
	//// Update - update multiple fields
	//db.Model(&product).Updates(Product{Price: 200, Code: "F42"}) // non-zero fields
	//db.Model(&product).Updates(map[string]interface{}{"Price": 200, "Code": "F42"})
	//
	//// Delete - delete product
	//db.Delete(&product, 1)
}

func UpdateSocket(addr string) {
	var setting Setting
	db.First(&setting)
	setting.Socket = addr
	if setting.ID < 0 {
		db.Create(&setting)
	}
	db.Save(&setting)
}

func UpdatePath(path string) {
	var setting Setting
	db.First(&setting)
	setting.Path = path
	if setting.ID < 0 {
		db.Create(&setting)
	}
	db.Save(&setting)
}

func GetConfig() (string, string) {
	var setting Setting
	db.First(&setting)
	if setting.ID > 0 {
		return setting.Socket, setting.Path
	}
	return "", ""
}

func NewRecord(nRecord Record) {
	db.Save(&nRecord)
}

func GetRecords() []Record {
	var records []Record
	db.Order("time desc").Find(&records)
	return records

}

func UpdateName(newName string, cTime int64) {
	db.Model(&Record{}).Where("time = ?", cTime).Update("name", newName)

	//if record.ID > 0 {
	//	record.Name = newName
	//	db.Save(&record)
	//}
}

func DelRecord(cTime int64) {
	db.Unscoped().Delete(&Record{}, "time = ?", cTime)

}
