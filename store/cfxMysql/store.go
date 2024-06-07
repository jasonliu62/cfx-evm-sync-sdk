package cfxMysql

import (
	"errors"
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/openweb3/web3go/types"
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

func StoreBlockAndTransactions(db *gorm.DB, block *types.Block, transactions []*types.TransactionDetail) error {
	return db.Transaction(func(tx *gorm.DB) error {
		dbBlock := ConvertBlockWithoutAuthor(block)
		authorName := ConvertAddressToString(block.Author)
		author, err := findOrCreateAddress(tx, authorName)
		if err != nil {
			return err
		}
		dbBlock.AuthorID = author.ID
		if err := tx.Create(&dbBlock).Error; err != nil {
			return fmt.Errorf("failed to create block: %w", err)
		}
		for _, transactionDetail := range transactions {
			dbTransactionDetail := ConvertTransactionDetail(transactionDetail)
			from, err := findOrCreateAddress(tx, ConvertAddressToString(&transactionDetail.From))
			if err != nil {
				return err
			}
			to, err := findOrCreateAddress(tx, ConvertAddressToString(transactionDetail.To))
			if err != nil {
				return err
			}
			dbTransactionDetail.FromAddress = from.ID
			dbTransactionDetail.ToAddress = to.ID
			if err := tx.Create(&dbTransactionDetail).Error; err != nil {
				return fmt.Errorf("failed to create transaction detail: %w", err)
			}
		}

		return nil
	})
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
