package notion

import (
	"fmt"
	"io"
	"sort"

	"github.com/gomarkdown/markdown/ast"
)

// RenderNodeFunc allows reusing most of Renderer logic and replacing
// rendering of some nodes. If it returns false, Renderer.RenderNode
// will execute its logic. If it returns true, Renderer.RenderNode will
// skip rendering this node and will return WalkStatus
type RenderNodeFunc func(w io.Writer, node ast.Node) (ast.WalkStatus, bool)

// RendererOptions is a collection of supplementary parameters tweaking
// the behavior of various parts of HTML renderer.
type RendererOptions struct {
	// if set, called at the start of RenderNode(). Allows replacing
	// rendering of some nodes
	RenderNodeHook RenderNodeFunc
}

// Renderer implements Renderer interface for HTML output.
//
// Do not create this directly, instead use the NewRenderer function.
type Renderer struct {
	opts            RendererOptions
	TableHolster    TableBlockBlock
	TableRowHolster TableRowBlockBlock
	BlocksFinal     []interface{}
	documentMatter  ast.DocumentMatters // keep track of front/main/back matter.
}

// NewRenderer creates and configures an Renderer object, which
// satisfies the Renderer interface.
func NewRenderer(opts RendererOptions) *Renderer {
	return &Renderer{
		opts: opts,
	}
}

// AddBlock appends a block to the holster
func (r *Renderer) AddBlock(inputBlock interface{}) {
	r.BlocksFinal = append(r.BlocksFinal, inputBlock)
}

// Text writes ast.Text node
func (r *Renderer) Text(w io.Writer, text *ast.Text) {
	textBlock, err := GetBlock[ParagraphBlock]("paragraph")
	if err != nil {
		fmt.Println("i can't do shit here")
	}
	nodeContentAsRichText := RichTextFromString(string(text.Literal))
	textBlock.Paragraph.RichText = append(textBlock.Paragraph.RichText, nodeContentAsRichText)
	r.AddBlock(textBlock)
}

// HardBreak writes ast.Hardbreak node
func (r *Renderer) HardBreak(w io.Writer, node *ast.Hardbreak) {
	textBlock, err := GetBlock[ParagraphBlock]("paragraph")
	if err != nil {
		fmt.Println("i can't do shit here")
	}
	nodeContentAsRichText := RichTextFromString(string("\n"))
	textBlock.Paragraph.RichText = append(textBlock.Paragraph.RichText, nodeContentAsRichText)
	r.AddBlock(textBlock)
}

// NonBlockingSpace writes ast.NonBlockingSpace node
func (r *Renderer) NonBlockingSpace(w io.Writer, node *ast.NonBlockingSpace) {
	textBlock, err := GetBlock[ParagraphBlock]("paragraph")
	if err != nil {
		fmt.Println("i can't do shit here")
	}
	nodeContentAsRichText := RichTextFromString(string(" "))
	textBlock.Paragraph.RichText = append(textBlock.Paragraph.RichText, nodeContentAsRichText)
	r.AddBlock(textBlock)
}

func (r *Renderer) Softbreak(w io.Writer, node *ast.Softbreak) {
	textBlock, err := GetBlock[ParagraphBlock]("paragraph")
	if err != nil {
		fmt.Println("i can't do shit here")
	}
	nodeContentAsRichText := RichTextFromString(string(" "))
	textBlock.Paragraph.RichText = append(textBlock.Paragraph.RichText, nodeContentAsRichText)
	r.AddBlock(textBlock)
}

// HTMLSpan writes ast.HTMLSpan node
func (r *Renderer) HTMLSpan(w io.Writer, span *ast.HTMLSpan) {
	textBlock, err := GetBlock[ParagraphBlock]("paragraph")
	if err != nil {
		fmt.Println("i can't do shit here")
	}
	nodeContentAsRichText := RichTextFromString(string(span.Literal))
	textBlock.Paragraph.RichText = append(textBlock.Paragraph.RichText, nodeContentAsRichText)
	r.AddBlock(textBlock)
}

// Link writes ast.Link node
func (r *Renderer) Link(w io.Writer, link *ast.Link) {
	textBlock, err := GetBlock[ParagraphBlock]("paragraph")
	if err != nil {
		fmt.Println("i can't do shit here")
	}
	nodeContentAsRichText := RichTextFromString(string(link.Literal))
	nodeContentAsRichText.Text.Link.URL = string(link.Destination)
	textBlock.Paragraph.RichText = append(textBlock.Paragraph.RichText, nodeContentAsRichText)
	r.AddBlock(textBlock)
}

// Image writes ast.Image node
func (r *Renderer) Image(w io.Writer, node *ast.Image) {
	textBlock, err := GetBlock[ImageBlockBlock]("image")
	if err != nil {
		fmt.Println("i can't do shit here")
	}
	textBlock.ImageBlock.External.URL = string(node.Destination)
	textBlock.ImageBlock.Caption = append(textBlock.ImageBlock.Caption, RichTextFromString(string(node.Literal)))
	r.AddBlock(textBlock)
}

// Paragraph writes ast.Paragraph node
func (r *Renderer) Paragraph(w io.Writer, para *ast.Paragraph) {
	textBlock, err := GetBlock[ParagraphBlock]("paragraph")
	if err != nil {
		fmt.Println("i can't do shit here")
	}
	nodeContentAsRichText := RichTextFromString(string(para.Literal))
	textBlock.Paragraph.RichText = append(textBlock.Paragraph.RichText, nodeContentAsRichText)
	r.AddBlock(textBlock)
}

// Code writes ast.Code node
func (r *Renderer) Code(w io.Writer, node *ast.Code) {
	textBlock, err := GetBlock[CodeBlock]("code")
	if err != nil {
		fmt.Println("i can't do shit here")
	}
	nodeContentAsRichText := RichTextFromString(string(node.Literal))
	nodeContentAsRichText.Annotations.code = true
	textBlock.Code.RichText = append(textBlock.Code.RichText, nodeContentAsRichText)
	r.AddBlock(textBlock)
}

func (r *Renderer) Math(w io.Writer, node *ast.Math) {
	textBlock, err := GetBlock[CodeBlock]("code")
	if err != nil {
		fmt.Println("i can't do shit here")
	}
	nodeContentAsRichText := RichTextFromString(string(node.Literal))
	nodeContentAsRichText.Annotations.code = true
	textBlock.Code.RichText = append(textBlock.Code.RichText, nodeContentAsRichText)
	r.AddBlock(textBlock)
}

func (r *Renderer) MathBlock(w io.Writer, node *ast.MathBlock) {
	textBlock, err := GetBlock[CodeBlock]("code")
	if err != nil {
		fmt.Println("i can't do shit here")
	}
	nodeContentAsRichText := RichTextFromString(string(node.Literal))
	nodeContentAsRichText.Annotations.code = true
	textBlock.Code.RichText = append(textBlock.Code.RichText, nodeContentAsRichText)
	r.AddBlock(textBlock)
}

// HTMLBlock write ast.HTMLBlock node
func (r *Renderer) HTMLBlock(w io.Writer, node *ast.HTMLBlock) {
	textBlock, err := GetBlock[ParagraphBlock]("paragraph")
	if err != nil {
		fmt.Println("i can't do shit here")
	}
	nodeContentAsRichText := RichTextFromString(string(node.Literal))
	textBlock.Paragraph.RichText = append(textBlock.Paragraph.RichText, nodeContentAsRichText)
	r.AddBlock(textBlock)
}

// Heading writes ast.Heading node
func (r *Renderer) Heading(w io.Writer, node *ast.Heading) {
	level := node.Level
	switch level {
	case 1:
		headerBlock, err := GetBlock[Heading1Block]("heading_1")
		if err != nil {
			fmt.Println("i can't do shit here")
		}
		textContent := RichTextFromString(string(node.Literal))
		textContent.Annotations.bold = true
		headerBlock.Heading1.RichText = append(headerBlock.Heading1.RichText, textContent)
		r.AddBlock(headerBlock)
	case 2:
		headerBlock, err := GetBlock[Heading2Block]("heading_2")
		if err != nil {
			fmt.Println("i can't do shit here")
		}
		textContent := RichTextFromString(string(node.Literal))
		textContent.Annotations.bold = true
		headerBlock.Heading2.RichText = append(headerBlock.Heading2.RichText, textContent)
		r.AddBlock(headerBlock)
	default:
		headerBlock, err := GetBlock[Heading3Block]("heading_3")
		if err != nil {
			fmt.Println("i can't do shit here")
		}
		textContent := RichTextFromString(string(node.Literal))
		textContent.Annotations.bold = true
		headerBlock.Heading3.RichText = append(headerBlock.Heading3.RichText, textContent)
		r.AddBlock(headerBlock)
	}
}

// HorizontalRule writes ast.HorizontalRule node
func (r *Renderer) HorizontalRule(w io.Writer, node *ast.HorizontalRule) {
	builtBlock, err := GetBlock[DividerBlockBlock]("divider")
	if err != nil {
		fmt.Println("i can't do shit here")
	}
	r.AddBlock(builtBlock)
}

func (r *Renderer) AddQuote(w io.Writer, node *ast.BlockQuote) {
	builtBlock, err := GetBlock[QuoteBlock]("quote")
	if err != nil {
		fmt.Println("i can't do shit here")
	}
	textContent := RichTextFromString(string(node.Literal))
	textContent.Annotations.underline = true
	builtBlock.Quote.RichText = append(builtBlock.Quote.RichText, textContent)
	r.AddBlock(builtBlock)
}

func (r *Renderer) AddEmp(w io.Writer, node *ast.Emph) {
	builtBlock, err := GetBlock[ParagraphBlock]("paragraph")
	if err != nil {
		fmt.Println("i can't do shit here")
	}
	textContent := RichTextFromString(string(node.Literal))
	textContent.Annotations.italic = true
	builtBlock.Paragraph.RichText = append(builtBlock.Paragraph.RichText, textContent)
	r.AddBlock(builtBlock)
}

func (r *Renderer) AddStrong(w io.Writer, node *ast.Strong) {
	builtBlock, err := GetBlock[ParagraphBlock]("paragraph")
	if err != nil {
		fmt.Println("i can't do shit here")
	}
	textContent := RichTextFromString(string(node.Literal))
	textContent.Annotations.bold = true
	builtBlock.Paragraph.RichText = append(builtBlock.Paragraph.RichText, textContent)
	r.AddBlock(builtBlock)
}

func (r *Renderer) AddDel(w io.Writer, node *ast.Del) {
	builtBlock, err := GetBlock[ParagraphBlock]("paragraph")
	if err != nil {
		fmt.Println("i can't do shit here")
	}
	textContent := RichTextFromString(string(node.Literal))
	textContent.Annotations.strikethrough = true
	builtBlock.Paragraph.RichText = append(builtBlock.Paragraph.RichText, textContent)
	r.AddBlock(builtBlock)
}

// List writes ast.List node
func (r *Renderer) List(w io.Writer, list *ast.List) {
	switch ast.ListType(list.ListFlags) {
	case ast.ListTypeOrdered:
		builtBlock, err := GetBlock[NumberedListItemBlock]("numbered_list_item")
		if err != nil {
			fmt.Println("i can't do shit here")
		}
		builtBlock.NumberedListItem.RichText = append(builtBlock.NumberedListItem.RichText, RichTextFromString(string(list.Literal)))
		r.AddBlock(builtBlock)
	default:
		builtBlock, err := GetBlock[BulletedListItemBlock]("bulleted_list_item")
		if err != nil {
			fmt.Println("i can't do shit here")
		}
		builtBlock.BulletedListItem.RichText = append(builtBlock.BulletedListItem.RichText, RichTextFromString(string(list.Literal)))
		r.AddBlock(builtBlock)
	}
}

// ListItem writes ast.ListItem node
func (r *Renderer) ListItem(w io.Writer, list *ast.ListItem) {
	switch ast.ListType(list.ListFlags) {
	case ast.ListTypeOrdered:
		builtBlock, err := GetBlock[NumberedListItemBlock]("numbered_list_item")
		if err != nil {
			fmt.Println("i can't do shit here")
		}
		builtBlock.NumberedListItem.RichText = append(builtBlock.NumberedListItem.RichText, RichTextFromString(string(list.Literal)))
		r.AddBlock(builtBlock)
	default:
		builtBlock, err := GetBlock[BulletedListItemBlock]("bulleted_list_item")
		if err != nil {
			fmt.Println("i can't do shit here")
		}
		builtBlock.BulletedListItem.RichText = append(builtBlock.BulletedListItem.RichText, RichTextFromString(string(list.Literal)))
		r.AddBlock(builtBlock)
	}
}

// CodeBlock writes ast.CodeBlock node
func (r *Renderer) CodeBlock(w io.Writer, codeBlock *ast.CodeBlock) {
	textBlock, err := GetBlock[CodeBlock]("code")
	if err != nil {
		fmt.Println("i can't do shit here")
	}
	nodeContentAsRichText := RichTextFromString(string(codeBlock.Literal))
	textBlock.Code.RichText = append(textBlock.Code.RichText, nodeContentAsRichText)
	r.AddBlock(textBlock)
}

// Caption writes ast.Caption node
func (r *Renderer) Caption(w io.Writer, caption *ast.Caption) {
}

// CaptionFigure writes ast.CaptionFigure node
func (r *Renderer) CaptionFigure(w io.Writer, figure *ast.CaptionFigure) {
}

// TableRow writes ast.TableRow node
func (r *Renderer) TableRow(w io.Writer, tableCell *ast.TableRow) {
	builtBlock, err := GetBlock[TableRowBlockBlock]("table_row")
	if err != nil {
		fmt.Println("i can't do shit here")
	}
	rowcount := len(builtBlock.TableRowBlock.Cells)
	for _, childNode := range tableCell.Children {
		builtCell := RichTextFromString(string(childNode.AsContainer().Literal))
		builtBlock.TableRowBlock.Cells[rowcount] = append(builtBlock.TableRowBlock.Cells[rowcount], builtCell)
	}
	r.AddBlock(builtBlock)
}

// TableBody writes ast.TableBody node
func (r *Renderer) TableBody(w io.Writer, node *ast.TableBody) {
	builtBlock, err := GetBlock[TableBlockBlock]("table")
	if err != nil {
		fmt.Println("i can't do shit here")
	}
	r.AddBlock(builtBlock)
}

// DocumentMatter writes ast.DocumentMatter
func (r *Renderer) DocumentMatter(w io.Writer, node *ast.DocumentMatter) {
	r.documentMatter = node.Matter
}

// Citation writes ast.Citation node
func (r *Renderer) Citation(w io.Writer, node *ast.Citation) {
	// for i, c := range node.Destination {
	// }
}

// Callout writes ast.Callout node
func (r *Renderer) Callout(w io.Writer, node *ast.Callout) {
	builtBlock, err := GetBlock[CalloutBlock]("callout")
	if err != nil {
		fmt.Println("i can't do shit here")
	}
	textContent := RichTextFromString(string(node.Literal))
	builtBlock.Callout.RichText = append(builtBlock.Callout.RichText, textContent)
	r.AddBlock(builtBlock)
}

func (r *Renderer) Aside(w io.Writer, node *ast.Aside) {
	builtBlock, err := GetBlock[CalloutBlock]("callout")
	if err != nil {
		fmt.Println("i can't do shit here")
	}
	textContent := RichTextFromString(string(node.Literal))
	builtBlock.Callout.RichText = append(builtBlock.Callout.RichText, textContent)
	r.AddBlock(builtBlock)
}

// Index writes ast.Index node
func (r *Renderer) Index(w io.Writer, node *ast.Index) {
	// there is no in-text representation.
}

// BlockAttrs takes a node and checks if it has block level attributes set. If so it
// will return a slice each containing a "key=value(s)" string.
func BlockAttrs(node ast.Node) []string {
	var attr *ast.Attribute
	if c := node.AsContainer(); c != nil && c.Attribute != nil {
		attr = c.Attribute
	}
	if l := node.AsLeaf(); l != nil && l.Attribute != nil {
		attr = l.Attribute
	}
	if attr == nil {
		return nil
	}

	var s []string
	if attr.ID != nil {
		s = append(s, fmt.Sprintf(`%s="%s"`, "id", attr.ID))
	}

	classes := ""
	for _, c := range attr.Classes {
		classes += " " + string(c)
	}
	if classes != "" {
		s = append(s, fmt.Sprintf(`class="%s"`, classes[1:])) // skip space we added.
	}

	// sort the attributes so it remain stable between runs
	var keys = []string{}
	for k := range attr.Attrs {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		s = append(s, fmt.Sprintf(`%s="%s"`, k, attr.Attrs[k]))
	}

	return s
}

// RenderNode renders a markdown node to HTML
func (r *Renderer) RenderNode(w io.Writer, node ast.Node) ast.WalkStatus {
	if r.opts.RenderNodeHook != nil {
		status, didHandle := r.opts.RenderNodeHook(w, node)
		if didHandle {
			return status
		}
	}
	switch node := node.(type) {
	case *ast.Text:
		r.Text(w, node)
	case *ast.Softbreak:
		r.Softbreak(w, node)
	case *ast.Hardbreak:
		r.HardBreak(w, node)
	case *ast.NonBlockingSpace:
		r.NonBlockingSpace(w, node)
	case *ast.Emph:
		r.AddEmp(w, node)
	case *ast.Strong:
		r.AddStrong(w, node)
	case *ast.Del:
		r.AddDel(w, node)
	case *ast.BlockQuote:
		r.AddQuote(w, node)
	case *ast.Aside:
		r.Aside(w, node)
	case *ast.Link:
		r.Link(w, node)
	case *ast.Citation:
		r.Citation(w, node)
	case *ast.Image:
		r.Image(w, node)
	case *ast.Code:
		r.Code(w, node)
	case *ast.CodeBlock:
		r.CodeBlock(w, node)
	case *ast.Caption:
		r.Caption(w, node)
	case *ast.CaptionFigure:
		r.CaptionFigure(w, node)
	case *ast.Paragraph:
		r.Paragraph(w, node)
	case *ast.HTMLSpan:
		r.HTMLSpan(w, node)
	case *ast.HTMLBlock:
		r.HTMLBlock(w, node)
	case *ast.Heading:
		r.Heading(w, node)
	case *ast.HorizontalRule:
		r.HorizontalRule(w, node)
	case *ast.List:
		r.List(w, node)
	case *ast.ListItem:
		r.ListItem(w, node)
	case *ast.TableBody:
		r.TableBody(w, node)
	case *ast.TableRow:
		r.TableRow(w, node)
	case *ast.Math:
		r.Math(w, node)
	case *ast.MathBlock:
		r.MathBlock(w, node)
	case *ast.DocumentMatter:
		r.DocumentMatter(w, node)
	case *ast.Callout:
		r.Callout(w, node)
	case *ast.Index:
		r.Index(w, node)
	default:
		panic(fmt.Sprintf("Unknown node %T", node))
	}
	return ast.GoToNext // forces forward momentum
}
