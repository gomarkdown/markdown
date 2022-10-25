package notion

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"sort"

	"github.com/gomarkdown/markdown/ast"
)

// RenderNodeFunc allows reusing most of Renderer logic and replacing
// rendering of some nodes. If it returns false, Renderer.RenderNode
// will execute its logic. If it returns true, Renderer.RenderNode will
// skip rendering this node and will return WalkStatus
type RenderNodeFunc func(w io.Writer, node ast.Node, render *Renderer) (ast.WalkStatus, bool)

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
	InternalWriter  io.Writer
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

// Render returns a formatted json string with the blocks array available under the key "blocks"
func Render(doc ast.Node, renderer Renderer) string {
	var buf bytes.Buffer
	ast.WalkFunc(doc, func(node ast.Node, entering bool) ast.WalkStatus {
		return renderer.RenderNode(&buf, node)
	})
	return fmt.Sprintf("{\n\"blocks\": [%s]\n}", buf.Bytes())
}

// Text writes ast.Text node
func (r *Renderer) Text(w io.Writer, text *ast.Text) {
	builtBlock, err := GetBlock[ParagraphBlock](ParagraphType)
	if err != nil {
		fmt.Printf("i can't do shit here: %v", err)
	}
	nodeContentAsRichText := RichTextFromString(string(text.Literal))
	builtBlock.Paragraph.RichText = append(builtBlock.Paragraph.RichText, nodeContentAsRichText)
	r.AddBlock(builtBlock)
	out, err := json.MarshalIndent(builtBlock, "", "  ")
	if err != nil {
		fmt.Printf("i can't do marshall here: %v", err)
	}
	io.WriteString(w, string(out)+",\n")
}

// HardBreak writes ast.Hardbreak node
func (r *Renderer) HardBreak(w io.Writer, node *ast.Hardbreak) {
	builtBlock, err := GetBlock[ParagraphBlock](ParagraphType)
	if err != nil {
		fmt.Printf("i can't do shit here: %v", err)
	}
	nodeContentAsRichText := RichTextFromString(string("\n"))
	builtBlock.Paragraph.RichText = append(builtBlock.Paragraph.RichText, nodeContentAsRichText)
	r.AddBlock(builtBlock)
	out, err := json.MarshalIndent(builtBlock, "", "  ")
	if err != nil {
		fmt.Printf("i can't do marshall here: %v", err)
	}
	io.WriteString(w, string(out)+",\n")
}

// NonBlockingSpace writes ast.NonBlockingSpace node
func (r *Renderer) NonBlockingSpace(w io.Writer, node *ast.NonBlockingSpace) {
	builtBlock, err := GetBlock[ParagraphBlock](ParagraphType)
	if err != nil {
		fmt.Printf("i can't do shit here: %v", err)
	}
	nodeContentAsRichText := RichTextFromString(string(" "))
	builtBlock.Paragraph.RichText = append(builtBlock.Paragraph.RichText, nodeContentAsRichText)
	r.AddBlock(builtBlock)
	out, err := json.MarshalIndent(builtBlock, "", "  ")
	if err != nil {
		fmt.Printf("i can't do marshall here: %v", err)
	}
	io.WriteString(w, string(out)+",\n")
}

func (r *Renderer) Softbreak(w io.Writer, node *ast.Softbreak) {
	builtBlock, err := GetBlock[ParagraphBlock](ParagraphType)
	if err != nil {
		fmt.Printf("i can't do shit here: %v", err)
	}
	nodeContentAsRichText := RichTextFromString(string(" "))
	builtBlock.Paragraph.RichText = append(builtBlock.Paragraph.RichText, nodeContentAsRichText)
	r.AddBlock(builtBlock)
	out, err := json.MarshalIndent(builtBlock, "", "  ")
	if err != nil {
		fmt.Printf("i can't do marshall here: %v", err)
	}
	io.WriteString(w, string(out)+",\n")
}

// HTMLSpan writes ast.HTMLSpan node
func (r *Renderer) HTMLSpan(w io.Writer, span *ast.HTMLSpan) {
	builtBlock, err := GetBlock[ParagraphBlock](ParagraphType)
	if err != nil {
		fmt.Printf("i can't do shit here: %v", err)
	}
	nodeContentAsRichText := RichTextFromString(string(span.Literal))
	builtBlock.Paragraph.RichText = append(builtBlock.Paragraph.RichText, nodeContentAsRichText)
	r.AddBlock(builtBlock)
	out, err := json.MarshalIndent(builtBlock, "", "  ")
	if err != nil {
		fmt.Printf("i can't do marshall here: %v", err)
	}
	io.WriteString(w, string(out)+",\n")
}

// Link writes ast.Link node
func (r *Renderer) Link(w io.Writer, link *ast.Link) {
	builtBlock, err := GetBlock[ParagraphBlock](ParagraphType)
	if err != nil {
		fmt.Printf("i can't do shit here: %v", err)
	}
	nodeContentAsRichText := RichTextFromString(string(link.Literal))
	nodeContentAsRichText.Text.Link.URL = string(link.Destination)
	builtBlock.Paragraph.RichText = append(builtBlock.Paragraph.RichText, nodeContentAsRichText)
	r.AddBlock(builtBlock)
}

// Image writes ast.Image node
func (r *Renderer) Image(w io.Writer, node *ast.Image) {
	builtBlock, err := GetBlock[ImageBlockBlock](ImageBlockType)
	if err != nil {
		fmt.Printf("i can't do shit here: %v", err)
	}
	builtBlock.ImageBlock.External.URL = string(node.Destination)
	builtBlock.ImageBlock.Caption = append(builtBlock.ImageBlock.Caption, RichTextFromString(string(node.Literal)))
	r.AddBlock(builtBlock)
	out, err := json.MarshalIndent(builtBlock, "", "  ")
	if err != nil {
		fmt.Printf("i can't do marshall here: %v", err)
	}
	io.WriteString(w, string(out)+",\n")
}

// Paragraph writes ast.Paragraph node
func (r *Renderer) Paragraph(w io.Writer, para *ast.Paragraph) {
	builtBlock, err := GetBlock[ParagraphBlock](ParagraphType)
	if err != nil {
		fmt.Printf("i can't do shit here: %v", err)
	}
	nodeContentAsRichText := RichTextFromString(string(para.Literal))
	builtBlock.Paragraph.RichText = append(builtBlock.Paragraph.RichText, nodeContentAsRichText)
	r.AddBlock(builtBlock)
	out, err := json.MarshalIndent(builtBlock, "", "  ")
	if err != nil {
		fmt.Printf("i can't do marshall here: %v", err)
	}
	io.WriteString(w, string(out)+",\n")
}

// Code writes ast.Code node
func (r *Renderer) Code(w io.Writer, node *ast.Code) {
	builtBlock, err := GetBlock[CodeBlock](CodeType)
	if err != nil {
		fmt.Printf("i can't do shit here: %v", err)
	}
	nodeContentAsRichText := RichTextFromString(string(node.Literal))
	nodeContentAsRichText.Annotations.code = true
	builtBlock.Code.RichText = append(builtBlock.Code.RichText, nodeContentAsRichText)
	r.AddBlock(builtBlock)
	out, err := json.MarshalIndent(builtBlock, "", "  ")
	if err != nil {
		fmt.Printf("i can't do marshall here: %v", err)
	}
	io.WriteString(w, string(out)+",\n")
}

func (r *Renderer) Math(w io.Writer, node *ast.Math) {
	builtBlock, err := GetBlock[CodeBlock](CodeType)
	if err != nil {
		fmt.Printf("i can't do shit here: %v", err)
	}
	nodeContentAsRichText := RichTextFromString(string(node.Literal))
	nodeContentAsRichText.Annotations.code = true
	builtBlock.Code.RichText = append(builtBlock.Code.RichText, nodeContentAsRichText)
	r.AddBlock(builtBlock)
	out, err := json.MarshalIndent(builtBlock, "", "  ")
	if err != nil {
		fmt.Printf("i can't do marshall here: %v", err)
	}
	io.WriteString(w, string(out)+",\n")
}

func (r *Renderer) MathBlock(w io.Writer, node *ast.MathBlock) {
	builtBlock, err := GetBlock[CodeBlock](CodeType)
	if err != nil {
		fmt.Printf("i can't do shit here: %v", err)
	}
	nodeContentAsRichText := RichTextFromString(string(node.Literal))
	nodeContentAsRichText.Annotations.code = true
	builtBlock.Code.RichText = append(builtBlock.Code.RichText, nodeContentAsRichText)
	r.AddBlock(builtBlock)
	out, err := json.MarshalIndent(builtBlock, "", "  ")
	if err != nil {
		fmt.Printf("i can't do marshall here: %v", err)
	}
	io.WriteString(w, string(out)+",\n")
}

// HTMLBlock write ast.HTMLBlock node
func (r *Renderer) HTMLBlock(w io.Writer, node *ast.HTMLBlock) {
	builtBlock, err := GetBlock[ParagraphBlock](ParagraphType)
	if err != nil {
		fmt.Printf("i can't do shit here: %v", err)
	}
	nodeContentAsRichText := RichTextFromString(string(node.Literal))
	builtBlock.Paragraph.RichText = append(builtBlock.Paragraph.RichText, nodeContentAsRichText)
	r.AddBlock(builtBlock)
	out, err := json.MarshalIndent(builtBlock, "", "  ")
	if err != nil {
		fmt.Printf("i can't do marshall here: %v", err)
	}
	io.WriteString(w, string(out)+",\n")
}

// Heading writes ast.Heading node
func (r *Renderer) Heading(w io.Writer, node *ast.Heading) {
	level := node.Level
	switch level {
	case 1:
		headerBlock, err := GetBlock[Heading1Block](Heading1Type)
		if err != nil {
			fmt.Printf("i can't do shit here: %v", err)
		}
		textContent := RichTextFromString(string(node.Literal))
		textContent.Annotations.bold = true
		headerBlock.Heading1.RichText = append(headerBlock.Heading1.RichText, textContent)
		r.AddBlock(headerBlock)
		out, err := json.MarshalIndent(headerBlock, "", "  ")
		if err != nil {
			fmt.Printf("i can't do marshall here: %v", err)
		}
		io.WriteString(w, string(out)+",\n")
	case 2:
		headerBlock, err := GetBlock[Heading2Block](Heading2Type)
		if err != nil {
			fmt.Printf("i can't do shit here: %v", err)
		}
		textContent := RichTextFromString(string(node.Literal))
		textContent.Annotations.bold = true
		headerBlock.Heading2.RichText = append(headerBlock.Heading2.RichText, textContent)
		r.AddBlock(headerBlock)
		out, err := json.MarshalIndent(headerBlock, "", "  ")
		if err != nil {
			fmt.Printf("i can't do marshall here: %v", err)
		}
		io.WriteString(w, string(out)+",\n")
	default:
		headerBlock, err := GetBlock[Heading3Block](Heading3Type)
		if err != nil {
			fmt.Printf("i can't do shit here: %v", err)
		}
		textContent := RichTextFromString(string(node.Literal))
		textContent.Annotations.bold = true
		headerBlock.Heading3.RichText = append(headerBlock.Heading3.RichText, textContent)
		r.AddBlock(headerBlock)
		out, err := json.MarshalIndent(headerBlock, "", "  ")
		if err != nil {
			fmt.Printf("i can't do marshall here: %v", err)
		}
		io.WriteString(w, string(out)+",\n")
	}
}

// HorizontalRule writes ast.HorizontalRule node
func (r *Renderer) HorizontalRule(w io.Writer, node *ast.HorizontalRule) {
	builtBlock, err := GetBlock[DividerBlockBlock](DividerBlockType)
	if err != nil {
		fmt.Printf("i can't do shit here: %v", err)
	}
	r.AddBlock(builtBlock)
	out, err := json.MarshalIndent(builtBlock, "", "  ")
	if err != nil {
		fmt.Printf("i can't do marshall here: %v", err)
	}
	io.WriteString(w, string(out)+",\n")
}

func (r *Renderer) AddQuote(w io.Writer, node *ast.BlockQuote) {
	builtBlock, err := GetBlock[QuoteBlock](QuoteType)
	if err != nil {
		fmt.Printf("i can't do shit here: %v", err)
	}
	textContent := RichTextFromString(string(node.Literal))
	textContent.Annotations.underline = true
	builtBlock.Quote.RichText = append(builtBlock.Quote.RichText, textContent)
	r.AddBlock(builtBlock)
	out, err := json.MarshalIndent(builtBlock, "", "  ")
	if err != nil {
		fmt.Printf("i can't do marshall here: %v", err)
	}
	io.WriteString(w, string(out)+",\n")
}

func (r *Renderer) AddEmp(w io.Writer, node *ast.Emph) {
	builtBlock, err := GetBlock[ParagraphBlock](ParagraphType)
	if err != nil {
		fmt.Printf("i can't do shit here: %v", err)
	}
	textContent := RichTextFromString(string(node.Literal))
	textContent.Annotations.italic = true
	builtBlock.Paragraph.RichText = append(builtBlock.Paragraph.RichText, textContent)
	r.AddBlock(builtBlock)
	out, err := json.MarshalIndent(builtBlock, "", "  ")
	if err != nil {
		fmt.Printf("i can't do marshall here: %v", err)
	}
	io.WriteString(w, string(out)+",\n")
}

func (r *Renderer) AddStrong(w io.Writer, node *ast.Strong) {
	builtBlock, err := GetBlock[ParagraphBlock](ParagraphType)
	if err != nil {
		fmt.Printf("i can't do shit here: %v", err)
	}
	textContent := RichTextFromString(string(node.Literal))
	textContent.Annotations.bold = true
	builtBlock.Paragraph.RichText = append(builtBlock.Paragraph.RichText, textContent)
	r.AddBlock(builtBlock)
	out, err := json.MarshalIndent(builtBlock, "", "  ")
	if err != nil {
		fmt.Printf("i can't do marshall here: %v", err)
	}
	io.WriteString(w, string(out)+",\n")
}

func (r *Renderer) AddDel(w io.Writer, node *ast.Del) {
	builtBlock, err := GetBlock[ParagraphBlock](ParagraphType)
	if err != nil {
		fmt.Printf("i can't do shit here: %v", err)
	}
	textContent := RichTextFromString(string(node.Literal))
	textContent.Annotations.strikethrough = true
	builtBlock.Paragraph.RichText = append(builtBlock.Paragraph.RichText, textContent)
	r.AddBlock(builtBlock)
	out, err := json.MarshalIndent(builtBlock, "", "  ")
	if err != nil {
		fmt.Printf("i can't do marshall here: %v", err)
	}
	io.WriteString(w, string(out)+",\n")
}

// List writes ast.List node
func (r *Renderer) List(w io.Writer, list *ast.List) {
	switch ast.ListType(list.ListFlags) {
	case ast.ListTypeOrdered:
		builtBlock, err := GetBlock[NumberedListItemBlock](NumberedListItemType)
		if err != nil {
			fmt.Printf("i can't do shit here: %v", err)
		}
		builtBlock.NumberedListItem.RichText = append(builtBlock.NumberedListItem.RichText, RichTextFromString(string(list.Literal)))
		r.AddBlock(builtBlock)
		out, err := json.MarshalIndent(builtBlock, "", "  ")
		if err != nil {
			fmt.Printf("i can't do marshall here: %v", err)
		}
		io.WriteString(w, string(out)+",\n")
	default:
		builtBlock, err := GetBlock[BulletedListItemBlock](BulletedListItemType)
		if err != nil {
			fmt.Printf("i can't do shit here: %v", err)
		}
		builtBlock.BulletedListItem.RichText = append(builtBlock.BulletedListItem.RichText, RichTextFromString(string(list.Literal)))
		r.AddBlock(builtBlock)
		out, err := json.MarshalIndent(builtBlock, "", "  ")
		if err != nil {
			fmt.Printf("i can't do marshall here: %v", err)
		}
		io.WriteString(w, string(out)+",\n")
	}
}

// ListItem writes ast.ListItem node
func (r *Renderer) ListItem(w io.Writer, list *ast.ListItem) {
	switch ast.ListType(list.ListFlags) {
	case ast.ListTypeOrdered:
		builtBlock, err := GetBlock[NumberedListItemBlock](NumberedListItemType)
		if err != nil {
			fmt.Printf("i can't do shit here: %v", err)
		}
		builtBlock.NumberedListItem.RichText = append(builtBlock.NumberedListItem.RichText, RichTextFromString(string(list.Literal)))
		r.AddBlock(builtBlock)
		out, err := json.MarshalIndent(builtBlock, "", "  ")
		if err != nil {
			fmt.Printf("i can't do marshall here: %v", err)
		}
		io.WriteString(w, string(out)+",\n")
	default:
		builtBlock, err := GetBlock[BulletedListItemBlock](BulletedListItemType)
		if err != nil {
			fmt.Printf("i can't do shit here: %v", err)
		}
		builtBlock.BulletedListItem.RichText = append(builtBlock.BulletedListItem.RichText, RichTextFromString(string(list.Literal)))
		r.AddBlock(builtBlock)
		out, err := json.MarshalIndent(builtBlock, "", "  ")
		if err != nil {
			fmt.Printf("i can't do marshall here: %v", err)
		}
		io.WriteString(w, string(out)+",\n")
	}
}

// CodeBlock writes ast.CodeBlock node
func (r *Renderer) CodeBlock(w io.Writer, codeBlock *ast.CodeBlock) {
	builtBlock, err := GetBlock[CodeBlock](CodeType)
	if err != nil {
		fmt.Printf("i can't do shit here: %v", err)
	}
	nodeContentAsRichText := RichTextFromString(string(codeBlock.Literal))
	builtBlock.Code.RichText = append(builtBlock.Code.RichText, nodeContentAsRichText)
	r.AddBlock(builtBlock)
	out, err := json.MarshalIndent(builtBlock, "", "  ")
	if err != nil {
		fmt.Printf("i can't do marshall here: %v", err)
	}
	io.WriteString(w, string(out)+",\n")
}

// Caption writes ast.Caption node
func (r *Renderer) Caption(w io.Writer, caption *ast.Caption) {
}

// CaptionFigure writes ast.CaptionFigure node
func (r *Renderer) CaptionFigure(w io.Writer, figure *ast.CaptionFigure) {
}

// TableRow writes ast.TableRow node
func (r *Renderer) TableRow(w io.Writer, tableCell *ast.TableRow) {
	builtBlock, err := GetBlock[TableRowBlockBlock](TableRowBlockType)
	if err != nil {
		fmt.Printf("i can't do shit here: %v", err)
	}
	rowcount := len(builtBlock.TableRowBlock.Cells)
	for _, childNode := range tableCell.Children {
		builtCell := RichTextFromString(string(childNode.AsContainer().Literal))
		builtBlock.TableRowBlock.Cells[rowcount] = append(builtBlock.TableRowBlock.Cells[rowcount], builtCell)
	}
	r.AddBlock(builtBlock)
	out, err := json.MarshalIndent(builtBlock, "", "  ")
	if err != nil {
		fmt.Printf("i can't do marshall here: %v", err)
	}
	io.WriteString(w, string(out)+",\n")
}

// TableBody writes ast.TableBody node
func (r *Renderer) TableBody(w io.Writer, node *ast.TableBody) {
	builtBlock, err := GetBlock[TableBlockBlock](TableBlockType)
	if err != nil {
		fmt.Printf("i can't do shit here: %v", err)
	}
	r.AddBlock(builtBlock)
	out, err := json.MarshalIndent(builtBlock, "", "  ")
	if err != nil {
		fmt.Printf("i can't do marshall here: %v", err)
	}
	io.WriteString(w, string(out)+",\n")
}

// DocumentMatter writes ast.DocumentMatter
func (r *Renderer) DocumentMatter(w io.Writer, node *ast.DocumentMatter) {
	r.documentMatter = node.Matter
}

// Citation writes ast.Citation node
func (r *Renderer) Citation(w io.Writer, node *ast.Citation) {
}

// Callout writes ast.Callout node
func (r *Renderer) Callout(w io.Writer, node *ast.Callout) {
	builtBlock, err := GetBlock[CalloutBlock](CalloutType)
	if err != nil {
		fmt.Printf("i can't do shit here: %v", err)
	}
	textContent := RichTextFromString(string(node.Literal))
	builtBlock.Callout.RichText = append(builtBlock.Callout.RichText, textContent)
	r.AddBlock(builtBlock)
	out, err := json.MarshalIndent(builtBlock, "", "  ")
	if err != nil {
		fmt.Printf("i can't do marshall here: %v", err)
	}
	io.WriteString(w, string(out)+",\n")
}

func (r *Renderer) Aside(w io.Writer, node *ast.Aside) {
	builtBlock, err := GetBlock[CalloutBlock](CalloutType)
	if err != nil {
		fmt.Printf("i can't do shit here: %v", err)
	}
	textContent := RichTextFromString(string(node.Literal))
	builtBlock.Callout.RichText = append(builtBlock.Callout.RichText, textContent)
	r.AddBlock(builtBlock)
	out, err := json.MarshalIndent(builtBlock, "", "  ")
	if err != nil {
		fmt.Printf("i can't do marshall here: %v", err)
	}
	io.WriteString(w, string(out)+",\n")
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
		status, didHandle := r.opts.RenderNodeHook(w, node, r)
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
