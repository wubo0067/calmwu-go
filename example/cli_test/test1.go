package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "greet"
	app.Usage = "fight the loneliness!"
	app.Action = func(c *cli.Context) error {
		fmt.Printf("Hello friend [%s]!\n", c.Args().Get(0))

		fmt.Println("Flag lang[%s]", c.String("lang"))
		fmt.Println("Flag count[%d]", c.Int("count"))
		return nil
	}

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "lang, l",
			Value: "chinaese",
			Usage: "language for the test1",
		},
		cli.IntFlag{
			Name:  "count, c",
			Value: 0,
			Usage: "count for the test1",
		},
	}

	app.Run(os.Args)
}
