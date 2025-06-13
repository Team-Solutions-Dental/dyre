package endpoint

import (
	"testing"
)

func TestTypescript(t *testing.T) {
	expectedTS := map[string]string{
		"Customers": `interface Customers { 
  CustomerID: string;
  FirstName?: string;
  LastName?: string;
  CreateDate?: Date;
  Active?: boolean;
  Zip?: number;
}`,
		"Invoices": `interface Invoices { 
  SaleID: string;
  Balance?: number;
  InvoiceNumber?: number;
  CreateDate?: Date;
}`,
		"Sales": `interface Sales { 
  CustomerID: string;
  SaleID?: number;
  CreateDate?: Date;
  Charge?: number;
}`,
	}
	service, err := ParseJSON([]byte(testingJSON()))
	if err != nil {
		t.Fatalf(`error: %v`, err)
	}

	for _, s := range service.Endpoints {
		exp, ok := expectedTS[s.Name]
		if !ok {
			t.Errorf("Field %s not found", s.Name)
			continue
		}
		if exp != s.TS() {
			t.Errorf("Expected %s TS output does not match\n%s\n%s\n", s.Name, exp, s.TS())
			continue
		}
	}

}
