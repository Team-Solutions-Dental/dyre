// package main
//
// import (
// 	"fmt"
// 	"log"
//
// 	"database/sql"
// 	"strings"
//
// 	"github.com/gin-gonic/gin"
// 	_ "github.com/mattn/go-sqlite3"
// 	"github.com/vamuscari/dyre"
// 	"gopkg.in/guregu/null.v4"
// )
//
// var Re map[string]dyre.DyReRequest
//
// // "gopkg.in/guregu/null.v4"
// func main() {
//
// 	var sqlTypes = map[string]interface{}{
// 		"null.String": null.String{},
// 		"null.Int":    null.Int{},
// 		"null.Bool":   null.Bool{},
// 		"null.Time":   null.Time{},
// 		"null.Float":  null.Float{},
// 	}
//
// 	// NOTE: sql types are needed for null ojects out of the db.
// 	dyre.AddTypes(sqlTypes)
// 	dyre.DefaultType = "null.String"
//
// 	var dyre_err error
// 	Re, dyre_err = dyre.Init("./dyre.json")
// 	if dyre_err != nil {
// 		log.Panicf("dyre init failed: %v", dyre_err)
// 	}
//
// 	routerURL := "127.0.0.1" + ":" + "4201"
//
// 	log.Println("Router URL:  " + routerURL)
//
// 	router := gin.Default()
// 	router.SetTrustedProxies([]string{"127.0.0.1"})
//
// 	router.GET("/Customers", getCustomers)
// 	router.GET("/Customers/:CustomerID", getCustomerByID)
//
// 	router.Run(routerURL)
// }
//
// func getCustomers(g *gin.Context) {
// 	fields_string, _ := g.GetQuery("fields")
// 	fields := strings.Split(fields_string, ",")
//
// 	groups_string, _ := g.GetQuery("groups")
// 	groups := strings.Split(groups_string, ",")
// 	customers, ok := Re["Customers"]
// 	if !ok {
// 		g.String(500, "Failed to start request")
// 		return
// 	}
//
// 	valid, err := customers.ValidateRequest(fields, groups)
// 	if err != nil {
// 		g.String(400, "Failed to parse request")
// 		return
// 	}
//
// 	query := fmt.Sprintf("SELECT %s FROM %s LIMIT 1000", strings.Join(valid.SQLFields(), ", "), customers.TableName())
//
// 	table, err := read_db(query, valid)
// 	if err != nil {
// 		g.String(500, "Failed to make request")
// 		return
// 	}
//
// 	output := make(map[string]any)
// 	headers := valid.Headers()
//
// 	output["Headers"] = headers
// 	output["Table"] = table
//
// 	g.JSON(200, output)
// 	return
// }
//
// func getCustomerByID(g *gin.Context) {
// 	id_string := g.Param("CustomerID")
// 	// ids := strings.Split(id_string, ",")
//
// 	fields_string, _ := g.GetQuery("fields")
// 	fields := strings.Split(fields_string, ",")
//
// 	groups_string, _ := g.GetQuery("groups")
// 	groups := strings.Split(groups_string, ",")
// 	customers, ok := Re["Customers"]
// 	if !ok {
// 		g.String(500, "Failed to start request")
// 		return
// 	}
//
// 	valid, err := customers.ValidateRequest(fields, groups)
// 	if err != nil {
// 		g.String(400, "Failed to parse request")
// 		return
// 	}
//
// 	query := fmt.Sprintf("SELECT %s FROM %s WHERE CustomerId = '%s'", strings.Join(valid.SQLFields(), ", "), customers.TableName(), id_string)
//
// 	table, err := read_db(query, valid)
// 	if err != nil {
// 		g.String(500, "Failed to make request")
// 		return
// 	}
//
// 	output := make(map[string]any)
// 	headers := valid.Headers()
//
// 	output["Headers"] = headers
// 	output["Table"] = table
//
// 	g.JSON(200, output)
// 	return
// }
//
// func read_db(query string, re dyre.DyRe_Validated) ([][]interface{}, error) {
//
// 	db, err := sql.Open("sqlite3", "./dyreExample.db")
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer db.Close()
//
// 	rows, err := db.Query(query)
// 	if err != nil {
// 		fmt.Printf("\nQuery Error: %v\nQuery:%v", err, query)
//
// 		return nil, err
// 	}
//
// 	var table [][]interface{}
//
// 	for rows.Next() {
//
// 		new_row := re.GenerateArray()
//
// 		err = rows.Scan(new_row...)
// 		if err != nil {
// 			fmt.Printf("Scan Error: %v", err)
// 			return nil, err
// 		}
// 		table = append(table, new_row)
// 	}
//
// 	return table, nil
// }
