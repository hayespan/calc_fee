package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"
)

type Data struct {
	Fees    [][]interface{} `json:"fees"`
	Tenants [][]interface{} `json:"tenants"`
}

type Tenant struct {
	Name      string
	StartDate time.Time
	EndDate   time.Time
	Total     float64
}
type TenantMap map[string]*Tenant

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: ./calc_fee <data_file>")
		return
	}
	filePath := os.Args[1]
	buf, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Println(err)
		return
	}

	data := &Data{}
	err = json.Unmarshal(buf, &data)
	if err != nil {
		fmt.Println(err)
		return
	}
	loc, _ := time.LoadLocation("Asia/Shanghai")

	parseDate := func(val interface{}) (time.Time, error) {
		startDateStr, err := val.(string)
		if !err {
			return time.Time{}, errors.New(fmt.Sprintf("wrong date value: +%v", val))
		}
		return time.ParseInLocation("2006-01-02", startDateStr, loc)
	}

	tenantMap := make(TenantMap)
	for _, item := range data.Tenants {
		if len(item) != 3 {
			fmt.Println("wrong data:", item)
			return
		}
		name, ok := item[0].(string)
		if !ok || name == "" {
			fmt.Println("wrong data:", item)
			return
		}
		startDate, err := parseDate(item[1])
		if err != nil {
			fmt.Println(err)
			return
		}
		endDate, err := parseDate(item[2])
		if err != nil {
			fmt.Println(err)
			return
		}
		tenantMap[name] = &Tenant{
			Name:      name,
			StartDate: startDate,
			EndDate:   endDate,
		}
	}

	total := .0
	// totalCheck := .0
	for _, item := range data.Fees {
		if len(item) != 4 {
			fmt.Println("wrong data:", item)
			return
		}
		startDate, err := parseDate(item[1])
		if err != nil {
			fmt.Println(err)
			return
		}
		endDate, err := parseDate(item[2])
		if err != nil {
			fmt.Println(err)
			return
		}
		fee, ok := item[3].(float64)
		if !ok {
			continue
		}

		dayCnt := int(endDate.Sub(startDate).Seconds() / 86400)
		feePerDay := fee / float64(dayCnt)
		for day := 0; day < dayCnt; day++ {
			t1 := startDate.Add(time.Duration(int64(time.Second) * int64(day) * 86400))
			t2 := startDate.Add(time.Duration(int64(time.Second) * int64(day+1) * 86400))

			hits := make([]*Tenant, 0, len(tenantMap))
			for _, tenant := range tenantMap {
				if (t1.After(tenant.StartDate) || t1.Equal(tenant.StartDate)) &&
					(t2.Before(tenant.EndDate) || t2.Equal(tenant.EndDate)) {
					hits = append(hits, tenant)
				}
			}
			if len(hits) == 0 {
				fmt.Println("wrong data: not covered fee.", item, t1, t2)
				return
			}

			feePerTenant := feePerDay / float64(len(hits))
			for _, tenant := range hits {
				tenant.Total += feePerTenant
				// totalCheck += feePerTenant
			}
		}

		total += fee
	}

	// fmt.Println("totalCheck:", totalCheck)
	for name, tenant := range tenantMap {
		fmt.Printf("%s: %.2f\n", name, tenant.Total)
	}
	fmt.Println("====")
	fmt.Printf("total: %.2f\n", total)
}
