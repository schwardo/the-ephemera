package main

import (
        "fmt"
	"strconv"
	"regexp"
	
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
	"gonum.org/v1/gonum/graph/path"
)

func Abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func CanOpen(room int) bool {
    rstr := strconv.Itoa(room)
    re, err := regexp.Compile("^([1-4](12|21|34|43)|(12|21|34|43)[1-4])$")
    if err != nil {
       panic(err)
    }
    return re.MatchString(rstr)
}

func IsAdjacent(room1 int, room2 int) bool {
    dx := Abs((room1 / 100) - (room2 / 100))
    dy := Abs(((room1 / 10) % 10) - ((room2 / 10) % 10))
    dz := Abs((room1 % 10) - (room2 % 10))

    if dx > 1 || dy > 1 || dz > 1 {
        return false;
    }
    return dx + dy + dz == 1;
}

func main() {
     g := simple.NewUndirectedGraph()
     nodes := make(map[int]graph.Node)

     dim := 4
     for x := 1; x <= dim; x++ {
          for y := 1; y <= dim; y++ {
     	      for z := 1; z <= dim; z++ {
	          room := 100*x + 10*y + z;
	      	  if !CanOpen(room) {
		     continue
		  }
		  n := g.NewNode()
		  g.AddNode(n)
		  nodes[room] = n

                  for r, o := range nodes {
		      if o != n && IsAdjacent(room, r) {
		          g.SetEdge(g.NewEdge(n, o))
		      }
		  }
	      }
	  }
     }

     lobby := nodes[124]

     ps := path.DijkstraAllPaths(g)
     for r, o := range nodes {
         if o != lobby {
             p, _, _ := ps.Between(lobby.ID(), o.ID())
             fmt.Printf("124 -> %d = %d\n", r, len(p)-1)
	 }
     }
}
