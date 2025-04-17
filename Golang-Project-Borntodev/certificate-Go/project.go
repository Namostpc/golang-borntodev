package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

const dbPath = "root:123456@tcp(localhost:3306)/products_db"

const productPath = "product"
const basePath = "/api"

type productInterface struct {
	Product_id    int     `json: "productid"`
	Product_name  string  `json: "productname"`
	Product_price float64 `json: "productprice"`
	Created_at    string  `json: "createat"`
}

func connectDB() (*sql.DB, error) {
	db, err := sql.Open("mysql", dbPath)
	if err != nil {
		log.Fatalf("Failed to connect to the Database %v :", err)
	} else {
		fmt.Println("Connected Successfully")
	}

	// fmt.Println("DB ===", db)
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	return db, nil
}

// Get all data in DB //
func getAllproduct(db *sql.DB) ([]productInterface, error) {
	ctx, cancle := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancle()

	result, err := db.QueryContext(ctx, `SELECT * FROM products`)

	if err != nil {
		log.Println(err.Error())
		return nil, err
	}
	defer result.Close()

	products := make([]productInterface, 0)
	for result.Next() {
		var productschema productInterface
		result.Scan(&productschema.Product_id,
			&productschema.Product_name,
			&productschema.Product_price,
			&productschema.Created_at)

		if err != nil {
			log.Println(err.Error())
			return nil, err
		}
		products = append(products, productschema)
	}
	fmt.Println("products ====", products)
	// fmt.Println("products ===", products)
	return products, nil
}

func getOnlyoneData(db *sql.DB, productId int) (*productInterface, error) {
	ctx, cancle := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancle()

	row := db.QueryRowContext(ctx, `SELECT * FROM products WHERE product_id = ?`, productId)

	productschem := &productInterface{}
	err := row.Scan(&productschem.Product_id,
		&productschem.Product_name,
		&productschem.Product_price,
		&productschem.Created_at)

	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return productschem, nil
}

// handler get all product //
func handlerAllProducts(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	switch r.Method {
	case http.MethodGet:
		productList, err := getAllproduct(db)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		j, err := json.Marshal(productList)

		if err != nil {
			log.Fatal(err)
		}

		_, err = w.Write(j)

		if err != nil {
			log.Fatal(err)
		}
	case http.MethodOptions:
		return
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)

	}
}

func handlerSingledata(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	urlPathSegment := strings.Split(r.URL.Path, fmt.Sprintf("%s/", productPath))
	if len(urlPathSegment[1:]) > 1 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	productid, err := strconv.Atoi(urlPathSegment[len(urlPathSegment)-1])
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		productdata, err := getOnlyoneData(db, productid)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if productdata == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		j, err := json.Marshal(productdata)

		if err != nil {
			log.Fatal(err)
		}

		_, err = w.Write(j)

		if err != nil {
			log.Fatal(err)
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}

func corsMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Access-Control-Allow-Origin", "*")
		w.Header().Add("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Methods", "GET,PUT,POST,DELET")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Authorization")
		handler.ServeHTTP(w, r)
	})
}

func setupRoutes(apiBasePath string, db *sql.DB) {
	allProductHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerAllProducts(w, r, db)
	})
	http.Handle(fmt.Sprintf("%s/%s", apiBasePath, productPath), corsMiddleware(allProductHandler))
	singleProductHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerSingledata(w, r, db)
	})
	http.Handle(fmt.Sprintf("%s/%s/", apiBasePath, productPath), corsMiddleware(singleProductHandler))

}

func main() {

	fmt.Println("Hello wotld")
	db, err := connectDB()
	if err != nil {
		log.Fatalf("Failed to connect %v", err)
	}
	defer db.Close()
	// getAllproduct(db)
	setupRoutes(basePath, db)
	log.Fatal(http.ListenAndServe(":3000", nil))

}
