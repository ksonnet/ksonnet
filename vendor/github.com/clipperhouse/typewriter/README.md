##What’s this?

Typewriter is a package to enable pluggable, type-driven codegen for Go. The envisioned use case is for generics-like functionality. This package forms the underpinning of [gen](https://github.com/clipperhouse/gen).

Usage is analogous to how codecs work with Go’s [image](http://golang.org/pkg/image/) package, or database drivers in the [sql](http://golang.org/pkg/database/sql/) package.

    import (
        // main package
    	“github.com/clipperhouse/typewriter”
    	
    	// any number of typewriters 
    	_ “github.com/clipperhouse/set”
    	_ “github.com/clipperhouse/linkedlist”
    )
    
    func main() {
    	app, err := typewriter.NewApp(”+gen”)
    	if err != nil {
    		panic(err)
    	}
    
    	app.WriteAll()
    }

Individual typewriters register themselves to the “parent” package via their init() functions. Have a look at one of the above typewriters to get an idea.
