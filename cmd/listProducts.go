/*
Copyright © 2019 Kondukto

*/

package cmd

import (
	"github.com/kondukto-io/kdt/client"
	"github.com/spf13/cobra"
)

// listProductsCmd represents the listProductsCmd command
var listProductsCmd = &cobra.Command{
	Use:   "product",
	Short: "lists products in Kondukto",
	Run:   productsRootCommand,
}

func init() {
	listCmd.AddCommand(listProductsCmd)

	listProductsCmd.Flags().StringP("name", "n", "", "search by name")
}

func productsRootCommand(cmd *cobra.Command, args []string) {
	c, err := client.New()
	if err != nil {
		qwe(ExitCodeError, err, "could not initialize Kondukto client")
	}

	name, err := cmd.Flags().GetString("name")
	if err != nil {
		qwe(ExitCodeError, err, "failed to get name flag")
	}

	products, err := c.ListProducts(name)
	if err != nil {
		qwe(ExitCodeError, err, "could not retrieve products")
	}

	if len(products) < 1 {
		qwm(ExitCodeError, "no products found")
	}

	productRows := []Row{
		{Columns: []string{"NAME", "ID", "PROJECTS COUNT", "UI Link"}},
		{Columns: []string{"----", "--", "--------------", "-------"}},
	}

	for _, project := range products {
		productRows = append(productRows, Row{Columns: project.FieldsAsRow()})
	}
	TableWriter(productRows...)
}
