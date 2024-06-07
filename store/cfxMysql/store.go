package cfxMysql

import (
	"errors"
	"fmt"
	"github.com/ghodss/yaml"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"os"
)

type Config struct {
	Database struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		Name     string `yaml:"name"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
	} `yaml:"database"`
}

func LoadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func NewDB(config *Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.Database.User, config.Database.Password, config.Database.Host, config.Database.Port, config.Database.Name)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
}

func InitDB(db *gorm.DB) error {
	err := db.AutoMigrate(&Block{}, &Address{}, &TransactionDetail{})
	if err != nil {
		return err
	}
	return nil
}

func StoreBlock(db *gorm.DB, block Block, authorName string) error {
	author, err := findOrCreateAddress(db, authorName)
	if err != nil {
		return err
	}
	block.AuthorID = author.ID
	return db.Create(&block).Error
}

func StoreTransactionDetail(db *gorm.DB, transactionDetail TransactionDetail, fromAddress, toAddress string) error {
	from, err := findOrCreateAddress(db, fromAddress)
	if err != nil {
		return err
	}
	to, err := findOrCreateAddress(db, toAddress)
	if err != nil {
		return err
	}
	transactionDetail.FromAddress = from.ID
	transactionDetail.ToAddress = to.ID
	return db.Create(&transactionDetail).Error
}

func findOrCreateAddress(db *gorm.DB, addressStr string) (Address, error) {
	var address Address
	if err := db.Where("address = ?", addressStr).First(&address).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Address not found, create a new address
			address = Address{Address: addressStr}
			if err = db.Create(&address).Error; err != nil {
				return address, fmt.Errorf("failed to create address: %w", err)
			}
		} else {
			return address, fmt.Errorf("failed to query address: %w", err)
		}
	}
	return address, nil
}

func Start() *gorm.DB {
	// Load configuration
	config, err := LoadConfig("./store/cfxMysql/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database
	db, err := NewDB(config)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	err = InitDB(db)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	return db

}
