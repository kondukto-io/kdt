/*
Copyright © 2021 Kondukto

*/

package cmd

import (
	"errors"
	"strings"

	"github.com/spf13/cobra"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/kondukto-io/kdt/client"
	"github.com/kondukto-io/kdt/klog"
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
	createProductCmd.Flags().StringP("projects", "p", "", "comma separated name or id of kondukto projects")
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
		if !primitive.IsValidObjectID(pr) {
			project, err := p.client.FindProjectByName(pr)
			if err != nil {
				klog.Debugf("failed to get [%s] project details: %v", pr, err)
				continue
			}
			pd.ID = project.ID
		} else {
			pd.ID = pr
		}

		parsedProjects = append(parsedProjects, pd)
	}

	name, err := p.cmd.Flags().GetString("name")
	if err != nil {
		qwe(ExitCodeError, err, "failed to parse the name flag: %v")
	}

	product, created := p.createProduct(name, parsedProjects)
	if created {
		qwm(ExitCodeSuccess, "product created successfully")
	}

	if len(parsedProjects) > 0 {
		p.updateProduct(product, parsedProjects)
		qwm(ExitCodeSuccess, "product updated successfully")
	}
	qwm(ExitCodeSuccess, "product already exists")
}

func (p *Product) createProduct(name string, projects []client.Project) (*client.Product, bool) {
	if len(p.printRows) == 0 {
		p.printRows = productPrintHeaders()
	}
	product, err := p.client.FindProductByName(name)
	if err != nil && !errors.Is(err, client.ProductNotFound) {
		qwe(ExitCodeError, err, "failed to get product by name: %v")
	}

	if product != nil {
		klog.Println("product found by given name parameter")
		return product, false
	}

	var pd = client.ProductDetail{
		Name:     name,
		Projects: projects,
	}

	product, err = p.client.CreateProduct(pd)
	if err != nil {
		qwe(ExitCodeError, err, "failed to create product")
	}

	p.printRows = append(p.printRows, Row{Columns: product.FieldsAsRow()})

	TableWriter(p.printRows...)
	klog.Printf("product [%s] created successfully", product.Name)
	return product, true
}

func (p *Product) updateProduct(product *client.Product, projects []client.Project) *client.Product {
	if len(p.printRows) == 0 {
		p.printRows = productPrintHeaders()
	}

	detail, err := p.client.GetProductDetail(product.ID)
	if err != nil {
		qwe(ExitCodeError, err, "failed to get product detail")
	}

	for _, pr := range projects {
		var add = func() bool {
			for _, pd := range detail.Projects {
				if pd.ID == pr.ID {
					return false
				}
			}
			return true
		}

		if add() {
			detail.Projects = append(detail.Projects, client.Project{ID: pr.ID})
		}
	}

	product, err = p.client.UpdateProduct(detail.ID, *detail)
	if err != nil {
		qwe(ExitCodeError, err, "failed to update product")
	}
	product.ProjectsCount = len(detail.Projects)

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
