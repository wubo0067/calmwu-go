/*
 * @Author: calmwu
 * @Date: 2019-12-02 15:24:55
 * @Last Modified by: calmwu
 * @Last Modified time: 2019-12-02 16:00:08
 */

package main

import "fmt"

//go:generate echo hello
//go:generate go run generate_test1.go
//go:generate  echo file=$GOFILE pkg=$GOPACKAGE
// go get -d github.com/appliedgo/generics

// go build .\generate_test1.go .\gen_moviesort.go

type Movie struct {
	Title string  // movie title
	Year  int     // movie release year
	Rate  float32 // movie rating
}

func sortByYearAsc(i, j *Movie) bool { return i.Year < j.Year }
func sortByYearDes(i, j *Movie) bool { return i.Year > j.Year }

func displayMovies(header string, movies []*Movie) {
	fmt.Println(header)
	for _, m := range movies {
		fmt.Printf("\t- %s (%d) R:%.1f\n", m.Title, m.Year, m.Rate)
	}
	fmt.Println()
}

func main() {
	movies := []*Movie{
		&Movie{"The 400 Blows", 1959, 8.1},
		&Movie{"La Haine", 1995, 8.1},
		&Movie{"The Godfather", 1972, 9.2},
		&Movie{"The Godfather: Part II", 1974, 9},
		&Movie{"Mafioso", 1962, 7.7}}

	msw := &MovieSliceWrap{
		Ssot:        movies,
		CompareFunc: sortByYearAsc,
	}

	msw.Sort()

	displayMovies("Movies sorted by year", movies)

	msw = &MovieSliceWrap{
		Ssot:        movies,
		CompareFunc: sortByYearDes,
	}

	msw.Sort()

	displayMovies("Movies sorted by year", movies)
}
