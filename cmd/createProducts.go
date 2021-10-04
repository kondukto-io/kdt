/*
Copyright © 2021 Kondukto

*/

package cmd

import (
	"strings"

	"github.com/kondukto-io/kdt/client"
	"github.com/kondukto-io/kdt/klog"
	"github.com/spf13/cobra"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// createProductCmd represents the create product command
var createProductCmd = &cobra.Command{
	Use:   "product",
	Short: "creates a new product on Kondukto",
	Run:   createProductsRootCommand,
}

func init() {
	createCmd.AddCommand(createProductCmd)

	createProductCmd.Flags().StringP("name", "n", "", "product name")
	createProductCmd.Flags().StringP("projects", "p", "", "comma separeted name or id of kondukto projects")
}

type Product struct {
	cmd       *cobra.Command
	client    *client.Client
	printRows []Row
}

func createProductsRootCommand(cmd *cobra.Command, _ []string) {
	c, err := client.New()
	if err != nil {
		qwe(ExitCodeError, err, "could not initialize Kondukto client")
	}

	var p = Product{
		cmd:       cmd,
		client:    c,
		printRows: productPrintHeaders(),
	}

	p.createProduct()
	qwm(ExitCodeSuccess, "product created successfully")
}

func (p *Product) createProduct() *client.Product {
	if len(p.printRows) == 0 {
		p.printRows = productPrintHeaders()
	}

	name, err := p.cmd.Flags().GetString("name")
	if err != nil {
		qwe(ExitCodeError, err, "failed to parse the namme flag: %v")
	}

	projects, err := p.cmd.Flags().GetString("projects")
	if err != nil {
		qwe(ExitCodeError, err, "failed to parse the projects flag")
	}

	parsedProjects := make([]client.Project, 0)
	pMap := make(map[string]bool)
	for _, pr := range strings.Split(projects, ",") {
		if pr == "" {
			continue
		}
		if exist, ok := pMap[pr]; ok && exist {
			continue
		}
		pMap[pr] = true
		pd := client.Project{}
		pd.ID, err = primitive.ObjectIDFromHex(pr)
		if err != nil {
			project, err := p.client.FindProjectByName(pr)
			if err != nil {
				klog.Debugf("failed to get [%s] project details: %v", pr, err)
				continue
			}
			pd.ID = project.ID
			if exist, ok := pMap[pd.ID.Hex()]; ok && exist {
				continue
			}
			pMap[pd.ID.Hex()] = true
		}
		parsedProjects = append(parsedProjects, pd)
	}

	var pd = client.ProductDetail{
		Name:     name,
		Projects: parsedProjects,
	}

	product, err := p.client.CreateProduct(pd)
	if err != nil {
		qwe(ExitCodeError, err, "failed to create product")
	}

	p.printRows = append(p.printRows, Row{Columns: product.FieldsAsRow()})

	TableWriter(p.printRows...)

	return product
}

func productPrintHeaders() []Row {
	return []Row{
		{Columns: []string{"NAME", "ID", "PROJECTS COUNT", "UI Link"}},
		{Columns: []string{"----", "--", "--------------", "-------"}},
	}
}
