package htmlutil

import (
	g "github.com/maragudk/gomponents"
	. "github.com/maragudk/gomponents/html"
)

func BasePage(title string, content ...g.Node) g.Node {
	return HTML(
		Head(
			TitleEl(g.Text(title)),
			StyleEl(g.Text(`
body {
	font-family: sans-serif;
	font-size: 1.1rem;
	padding: 1em;
}
`)),
		),
		Body(content...),
	)
}

func UnorderedList(x []string) g.Node {
	return Ul(g.Map(x, func(s string) g.Node {
		return Li(g.Text(s))
	})...)
}
