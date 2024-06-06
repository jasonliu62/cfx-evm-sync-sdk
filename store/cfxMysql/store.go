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
	err := db.AutoMigrate(&Block{})
	if err != nil {
		return err
	}
	err = db.AutoMigrate(&Address{})
	if err != nil {
		return err
	}
	err = db.AutoMigrate(&TransactionDetail{})
	if err != nil {
		return err
	}
	return nil
}

func StoreBlock(db *gorm.DB, block Block, authorName string) error {
	var author Address

	// Check if the author exists
	if err := db.Where("address = ?", authorName).First(&author).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Author not found, create a new author
			author = Address{Address: authorName}
			if err = db.Create(&author).Error; err != nil {
				return fmt.Errorf("failed to create author: %w", err)
			}
		} else {
			return fmt.Errorf("failed to query author: %w", err)
		}
	}
	block.AuthorID = author.ID
	return db.Create(&block).Error
}

func GetBlockHashByNumber(db *gorm.DB, blockNumber uint) (string, error) {
	var block Block
	result := db.Where("block_number = ?", blockNumber).First(&block)
	if result.Error != nil {
		return "", result.Error
	}
	return block.Hash, nil
}

func StoreTransactionDetail(db *gorm.DB, transactionDetail TransactionDetail, fromAdress, toAdress, TxHash string) error {
	var (
		from Address
		to   Address
		hash Hash
	)
	if err := db.Where("address = ?", fromAdress).First(&from).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Author not found, create a new author
			from = Address{Address: fromAdress}
			if err = db.Create(&from).Error; err != nil {
				return fmt.Errorf("failed to create fromAddress: %w", err)
			}
		} else {
			return fmt.Errorf("failed to query fromAddress: %w", err)
		}
	}
	if err := db.Where("address = ?", toAdress).First(&to).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Author not found, create a new author
			to = Address{Address: toAdress}
			if err = db.Create(&to).Error; err != nil {
				return fmt.Errorf("failed to create toAddress: %w", err)
			}
		} else {
			return fmt.Errorf("failed to query toAddress: %w", err)
		}
	}
	if err := db.Where("hash = ?", TxHash).First(&hash).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Author not found, create a new author
			hash = Hash{Hash: TxHash}
			if err = db.Create(&hash).Error; err != nil {
				return fmt.Errorf("failed to create Hash: %w", err)
			}
		} else {
			return fmt.Errorf("failed to query Hash: %w", err)
		}
	}
	transactionDetail.FromAddress = from.ID
	transactionDetail.ToAddress = to.ID
	transactionDetail.TxHash = hash.ID
	return db.Create(&transactionDetail).Error
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
