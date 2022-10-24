package notion

import (
	"errors"
	"time"
)

// subsections of a request
type AnnotationResponse struct {
	bold          bool
	italic        bool
	strikethrough bool
	underline     bool
	code          bool
	Color         string `json:"color"`
}

// string type alias
type BlockID string
type BlockType string
type ApiColor string
type DataBaseID string
type PageID string
type WorkSpace bool

type BlockBase struct {
	Type           string            `json:"type"`
	ID             BlockID           `json:"id,omitempty"`
	CreatedTime    *time.Time        `json:"created_time,omitempty"`
	LastEditedTime *time.Time        `json:"last_edited_time,omitempty"`
	CreatedBy      map[string]string `json:"created_by,omitempty"`
	LastEditedBy   map[string]string `json:"last_edited_by,omitempty"`
	HasChildren    bool              `json:"has_children,omitempty"`
	Archived       bool              `json:"archived,omitempty"`
}

// interface to infer structs bb
type BlockHolster interface {
	TypeInternal() string
	ActualBlock() interface{}
}

// actual blocks
type RichTextTextContent struct {
	Content string `json:"content"`
	Link    struct {
		URL string `json:"url"`
	} `json:"link"`
}

const RichTextType string = "rich_text"

type RichText struct {
	BlockBase
	Text        RichTextTextContent `json:"rich_text"`
	PlaintText  string              `json:"plain_text"`
	Annotations AnnotationResponse  `json:"annotations"`
	HREF        string              `json:"href"`
}

func (rt RichText) TypeInternal() string {
	return RichTextType
}
func (rt RichText) ActualBlock() interface{} {
	return rt
}

func RichTextFromString(inputValue string) RichText {
	internalRT := RichText{}
	content := RichTextTextContent{}
	content.Content = inputValue
	internalRT.Type = "rich_text"
	internalRT.Text = content
	return internalRT
}

type Paragraph struct {
	RichText []RichText `json:"rich_text"`
	Color    ApiColor   `json:"color"`
}

const ParagraphType string = "paragraph"

type ParagraphBlock struct {
	BlockBase
	Paragraph Paragraph    `json:"paragraph"`
	Parent    BlockHolster `json:"parent"`
	Object    BlockType    `json:"object"`
}

func (pb ParagraphBlock) TypeInternal() string {
	return ParagraphType
}
func (pb ParagraphBlock) ActualBlock() interface{} {
	return pb
}

// parent block stubs
type DataBaseBlock struct {
	BlockBase
	DataBase DataBaseID `json:"database_id"`
}

type PageBlock struct {
	BlockBase
	PageID PageID `json:"page_id"`
}

type BlockBlock struct {
	BlockBase
	BlockID BlockID `json:"block_id"`
}

type WorkspaceBlock struct {
	BlockBase
	WorkSpace WorkSpace `json:"workspace"`
}

// all supported blocks
type Heading1 struct {
	RichText []RichText `json:"rich_text"`
	Color    ApiColor   `json:"color"`
}

const Heading1Type string = "heading_1"

type Heading1Block struct {
	BlockBase
	Heading1 Heading1     `json:"heading_1"`
	Parent   BlockHolster `json:"parent"`
	Object   BlockType    `json:"object"`
}

func (bI Heading1Block) TypeInternal() string {
	return Heading1Type
}
func (bI Heading1Block) ActualBlock() interface{} {
	return bI
}

type Heading2 struct {
	RichText []RichText `json:"rich_text"`
	Color    ApiColor   `json:"color"`
}

const Heading2Type string = "heading_2"

type Heading2Block struct {
	BlockBase
	Heading2 Heading2     `json:"heading_2"`
	Parent   BlockHolster `json:"parent"`
	Object   BlockType    `json:"object"`
}

func (bI Heading2Block) TypeInternal() string {
	return Heading2Type
}
func (bI Heading2Block) ActualBlock() interface{} {
	return bI
}

type Heading3 struct {
	RichText []RichText `json:"rich_text"`
	Color    ApiColor   `json:"color"`
}

const Heading3Type string = "heading_3"

type Heading3Block struct {
	BlockBase
	Heading3 Heading3     `json:"heading_3"`
	Parent   BlockHolster `json:"parent"`
	Object   BlockType    `json:"object"`
}

func (bI Heading3Block) TypeInternal() string {
	return Heading3Type
}
func (bI Heading3Block) ActualBlock() interface{} {
	return bI
}

type BulletedListItem struct {
	RichText []RichText `json:"rich_text"`
	Color    ApiColor   `json:"color"`
}

const BulletedListItemType string = "bulleted_list_item"

type BulletedListItemBlock struct {
	BlockBase
	BulletedListItem BulletedListItem `json:"bulleted_list_item"`
	Parent           BlockHolster     `json:"parent"`
	Object           BlockType        `json:"object"`
}

func (bI BulletedListItemBlock) TypeInternal() string {
	return BulletedListItemType
}
func (bI BulletedListItemBlock) ActualBlock() interface{} {
	return bI
}

type NumberedListItem struct {
	RichText []RichText `json:"rich_text"`
	Color    ApiColor   `json:"color"`
}

const NumberedListItemType string = "numbered_list_item"

type NumberedListItemBlock struct {
	BlockBase
	NumberedListItem NumberedListItem `json:"numbered_list_item"`
	Parent           BlockHolster     `json:"parent"`
	Object           BlockType        `json:"object"`
}

func (bI NumberedListItemBlock) TypeInternal() string {
	return NumberedListItemType
}
func (bI NumberedListItemBlock) ActualBlock() interface{} {
	return bI
}

type Quote struct {
	RichText []RichText `json:"rich_text"`
	Color    ApiColor   `json:"color"`
}

const QuoteType string = "quote"

type QuoteBlock struct {
	BlockBase
	Quote  Quote        `json:"quote"`
	Parent BlockHolster `json:"parent"`
	Object BlockType    `json:"object"`
}

func (bI QuoteBlock) TypeInternal() string {
	return QuoteType
}
func (bI QuoteBlock) ActualBlock() interface{} {
	return bI
}

type ToDo struct {
	RichText []RichText `json:"rich_text"`
	Color    ApiColor   `json:"color"`
	Checked  bool       `json:"checked"`
}

const ToDoType string = "to_do"

type ToDoBlock struct {
	BlockBase
	ToDo   ToDo         `json:"to_do"`
	Parent BlockHolster `json:"parent"`
	Object BlockType    `json:"object"`
}

func (bI ToDoBlock) TypeInternal() string {
	return ToDoType
}
func (bI ToDoBlock) ActualBlock() interface{} {
	return bI
}

type Toggle struct {
	RichText []RichText `json:"rich_text"`
	Color    ApiColor   `json:"color"`
}

const ToggleType string = "toggle"

type ToggleBlock struct {
	BlockBase
	Toggle Toggle       `json:"toggle"`
	Parent BlockHolster `json:"parent"`
	Object BlockType    `json:"object"`
}

func (bI ToggleBlock) TypeInternal() string {
	return ToggleType
}
func (bI ToggleBlock) ActualBlock() interface{} {
	return bI
}

type Template struct {
	RichText []RichText `json:"rich_text"`
}

const TemplateType string = "template"

type TemplateBlock struct {
	BlockBase
	Template Template     `json:"template"`
	Parent   BlockHolster `json:"parent"`
	Object   BlockType    `json:"object"`
}

func (bI TemplateBlock) TypeInternal() string {
	return TemplateType
}
func (bI TemplateBlock) ActualBlock() interface{} {
	return bI
}

type SyncedBlock struct {
	SyncedFrom struct {
		Type    string `json:"type"`
		BlockID string `json:"block_id"`
	} `json:"synced_from"`
}

const SyncedBlockType string = "synced_block"

type SyncedBlockBlock struct {
	BlockBase
	SyncedBlock SyncedBlock  `json:"synced_block"`
	Parent      BlockHolster `json:"parent"`
	Object      BlockType    `json:"object"`
}

func (bI SyncedBlockBlock) TypeInternal() string {
	return SyncedBlockType
}
func (bI SyncedBlockBlock) ActualBlock() interface{} {
	return bI
}

type ChildPage struct {
	Title string `json:"title"`
}

const ChildPageType string = "child_page"

type ChildPageBlock struct {
	BlockBase
	ChildPage ChildPage    `json:"child_page"`
	Parent    BlockHolster `json:"parent"`
	Object    BlockType    `json:"object"`
}

func (bI ChildPageBlock) TypeInternal() string {
	return ChildPageType
}
func (bI ChildPageBlock) ActualBlock() interface{} {
	return bI
}

type ChildDatabase struct {
	Title string `json:"title"`
}

const ChildDatabaseType string = "child_database"

type ChildDatabaseBlock struct {
	BlockBase
	ChildDatabase ChildDatabase `json:"child_database"`
	Parent        BlockHolster  `json:"parent"`
	Object        BlockType     `json:"object"`
}

func (bI ChildDatabaseBlock) TypeInternal() string {
	return ChildDatabaseType
}
func (bI ChildDatabaseBlock) ActualBlock() interface{} {
	return bI
}

type Equation struct {
	Expression string `json:"expression"`
}

const EquationType string = "equation"

type EquationBlock struct {
	BlockBase
	Equation Equation     `json:"equation"`
	Parent   BlockHolster `json:"parent"`
	Object   BlockType    `json:"object"`
}

func (bI EquationBlock) TypeInternal() string {
	return EquationType
}
func (bI EquationBlock) ActualBlock() interface{} {
	return bI
}

type Code struct {
	RichText []RichText `json:"rich_text"`
	Color    ApiColor   `json:"color"`
	Language string     `json:"language"`
}

const CodeType string = "code"

type CodeBlock struct {
	BlockBase
	Code   Code         `json:"code"`
	Parent BlockHolster `json:"parent"`
	Object BlockType    `json:"object"`
}

func (bI CodeBlock) TypeInternal() string {
	return CodeType
}
func (bI CodeBlock) ActualBlock() interface{} {
	return bI
}

type Callout struct {
	RichText []RichText  `json:"rich_text"`
	color    ApiColor    `json:"color"`
	Icon     interface{} `json:"icon"`
}

const CalloutType string = "callout"

type CalloutBlock struct {
	BlockBase
	Callout Callout      `json:"callout"`
	Parent  BlockHolster `json:"parent"`
	Object  BlockType    `json:"object"`
}

func (bI CalloutBlock) TypeInternal() string {
	return CalloutType
}
func (bI CalloutBlock) ActualBlock() interface{} {
	return bI
}

type DividerBlock struct {
}

const DividerBlockType string = "divider"

type DividerBlockBlock struct {
	BlockBase
	DividerBlock DividerBlock `json:"divider"`
	Parent       BlockHolster `json:"parent"`
	Object       BlockType    `json:"object"`
}

func (bI DividerBlockBlock) TypeInternal() string {
	return DividerBlockType
}
func (bI DividerBlockBlock) ActualBlock() interface{} {
	return bI
}

type Breadcrumb struct {
}

const BreadcrumbType string = "breadcrumb"

type BreadcrumbBlock struct {
	BlockBase
	Breadcrumb Breadcrumb   `json:"breadcrumb"`
	Parent     BlockHolster `json:"parent"`
	Object     BlockType    `json:"object"`
}

func (pb BreadcrumbBlock) TypeInternal() string {
	return BreadcrumbType
}
func (pb BreadcrumbBlock) ActualBlock() interface{} {
	return pb
}

type TableOfContentsBlock struct {
	color ApiColor `json:"color"`
}

const TableOfContentsBlockType string = "table_of_contents"

type TableOfContentsBlockBlock struct {
	BlockBase
	TableOfContentsBlock TableOfContentsBlock `json:"table_of_contents"`
	Parent               BlockHolster         `json:"parent"`
	Object               BlockType            `json:"object"`
}

func (pb TableOfContentsBlockBlock) TypeInternal() string {
	return TableOfContentsBlockType
}
func (pb TableOfContentsBlockBlock) ActualBlock() interface{} {
	return pb
}

type ColumnListBlock struct {
}

const ColumnListBlockType string = "column_list"

type ColumnListBlockBlock struct {
	BlockBase
	ColumnListBlock ColumnListBlock `json:"column_list"`
	Parent          BlockHolster    `json:"parent"`
	Object          BlockType       `json:"object"`
}

func (pb ColumnListBlockBlock) TypeInternal() string {
	return ColumnListBlockType
}
func (pb ColumnListBlockBlock) ActualBlock() interface{} {
	return pb
}

type ColumnBlock struct {
}

const ColumnBlockType string = "column"

type ColumnBlockBlock struct {
	BlockBase
	ColumnBlock ColumnBlock  `json:"column"`
	Parent      BlockHolster `json:"parent"`
	Object      BlockType    `json:"object"`
}

func (pb ColumnBlockBlock) TypeInternal() string {
	return ColumnBlockType
}
func (pb ColumnBlockBlock) ActualBlock() interface{} {
	return pb
}

type LinkToPageBlock struct {
	Type       string `json:"type"`
	PageID     string `json:"page_id"`
	DatabaseID string `json:"database_id"`
	CommentID  string `json:"comment_id"`
}

const LinkToPageBlockType string = "link_to_page"

type LinkToPageBlockBlock struct {
	BlockBase
	LinkToPageBlock LinkToPageBlock `json:"link_to_page"`
	Parent          BlockHolster    `json:"parent"`
	Object          BlockType       `json:"object"`
}

func (pb LinkToPageBlockBlock) TypeInternal() string {
	return LinkToPageBlockType
}
func (pb LinkToPageBlockBlock) ActualBlock() interface{} {
	return pb
}

type TableBlock struct {
	HasColumnHeader bool
	HasRowHeader    bool
	TableWidth      bool
}

const TableBlockType string = "table"

type TableBlockBlock struct {
	BlockBase
	TableBlock TableBlock   `json:"table"`
	Parent     BlockHolster `json:"parent"`
	Object     BlockType    `json:"object"`
}

func (pb TableBlockBlock) TypeInternal() string {
	return TableBlockType
}
func (pb TableBlockBlock) ActualBlock() interface{} {
	return pb
}

type TableRowBlock struct {
	Cells [][]RichText
}

const TableRowBlockType string = "table_row"

type TableRowBlockBlock struct {
	BlockBase
	TableRowBlock TableRowBlock `json:"table_row"`
	Parent        BlockHolster  `json:"parent"`
	Object        BlockType     `json:"object"`
}

func (pb TableRowBlockBlock) TypeInternal() string {
	return TableRowBlockType
}
func (pb TableRowBlockBlock) ActualBlock() interface{} {
	return pb
}

type EmbedBlock struct {
	URL     string
	Caption []RichText
}

const EmbedBlockType string = "embed"

type EmbedBlockBlock struct {
	BlockBase
	EmbedBlock EmbedBlock   `json:"embed"`
	Parent     BlockHolster `json:"parent"`
	Object     BlockType    `json:"object"`
}

func (pb EmbedBlockBlock) TypeInternal() string {
	return EmbedBlockType
}
func (pb EmbedBlockBlock) ActualBlock() interface{} {
	return pb
}

type BookmarkBlock struct {
	URL     string
	Caption []RichText
}

const BookmarkBlockType string = "bookmark"

type BookmarkBlockBlock struct {
	BlockBase
	BookmarkBlock BookmarkBlock `json:"bookmark"`
	Parent        BlockHolster  `json:"parent"`
	Object        BlockType     `json:"object"`
}

func (pb BookmarkBlockBlock) TypeInternal() string {
	return BookmarkBlockType
}
func (pb BookmarkBlockBlock) ActualBlock() interface{} {
	return pb
}

type ImageBlock struct {
	Type string `json:"type"`
	File struct {
		URL        string     `json:"url"`
		ExpiryTime *time.Time `json:"expiry_time"`
	} `json:"file"`
	External struct {
		URL string `json:"url"`
	} `json:"external"`
	Caption []RichText `json:"caption"`
}

const ImageBlockType string = "image"

type ImageBlockBlock struct {
	BlockBase
	ImageBlock ImageBlock   `json:"image"`
	Parent     BlockHolster `json:"parent"`
	Object     BlockType    `json:"object"`
}

func (pb ImageBlockBlock) TypeInternal() string {
	return ImageBlockType
}
func (pb ImageBlockBlock) ActualBlock() interface{} {
	return pb
}

type VideBlock struct {
	Type string
	File struct {
		URL        string
		ExpiryTime *time.Time
	}
	External struct {
		URL string
	}
	Caption []RichText
}

const VideBlockType string = "video"

type VideBlockBlock struct {
	BlockBase
	VideBlock VideBlock    `json:"video"`
	Parent    BlockHolster `json:"parent"`
	Object    BlockType    `json:"object"`
}

func (pb VideBlockBlock) TypeInternal() string {
	return VideBlockType
}
func (pb VideBlockBlock) ActualBlock() interface{} {
	return pb
}

type PdfBlock struct {
	Type string
	File struct {
		URL        string
		ExpiryTime *time.Time
	}
	External struct {
		URL string
	}
	Caption []RichText
}

const PdfBlockType string = "pdf"

type PdfBlockBlock struct {
	BlockBase
	PdfBlock PdfBlock     `json:"pdf"`
	Parent   BlockHolster `json:"parent"`
	Object   BlockType    `json:"object"`
}

func (pb PdfBlockBlock) TypeInternal() string {
	return PdfBlockType
}
func (pb PdfBlockBlock) ActualBlock() interface{} {
	return pb
}

type FileBlock struct {
	Type string
	File struct {
		URL        string
		ExpiryTime *time.Time
	}
	External struct {
		URL string
	}
	Caption []RichText
}

const FileBlockType string = "file"

type FileBlockBlock struct {
	BlockBase
	FileBlock FileBlock    `json:"file"`
	Parent    BlockHolster `json:"parent"`
	Object    BlockType    `json:"object"`
}

func (pb FileBlockBlock) TypeInternal() string {
	return FileBlockType
}
func (pb FileBlockBlock) ActualBlock() interface{} {
	return pb
}

type AudioBlock struct {
	Type string
	File struct {
		URL        string
		ExpiryTime *time.Time
	}
	External struct {
		URL string
	}
	Caption []RichText
}

const AudioBlockType string = "audio"

type AudioBlockBlock struct {
	BlockBase
	AudioBlock AudioBlock   `json:"audio"`
	Parent     BlockHolster `json:"parent"`
	Object     BlockType    `json:"object"`
}

func (pb AudioBlockBlock) TypeInternal() string {
	return AudioBlockType
}
func (pb AudioBlockBlock) ActualBlock() interface{} {
	return pb
}

type LinkPreviewBlock struct {
	URL string
}

const LinkPreviewBlockType string = "link_preview"

type LinkPreviewBlockBlock struct {
	BlockBase
	LinkPreviewBlock LinkPreviewBlock `json:"link_preview"`
	Parent           BlockHolster     `json:"parent"`
	Object           BlockType        `json:"object"`
}

func (pb LinkPreviewBlockBlock) TypeInternal() string {
	return LinkPreviewBlockType
}
func (pb LinkPreviewBlockBlock) ActualBlock() interface{} {
	return pb
}

type UnsupportedBlock struct {
}

const UnsupportedBlockType string = "unsupported"

type UnsupportedBlockBlock struct {
	BlockBase
	UnsupportedBlock UnsupportedBlock `json:"unsupported"`
	Parent           BlockHolster     `json:"parent"`
	Object           BlockType        `json:"object"`
}

func (pb UnsupportedBlockBlock) TypeInternal() string {
	return UnsupportedBlockType
}
func (pb UnsupportedBlockBlock) ActualBlock() interface{} {
	return pb
}

type AnyBlock interface {
	ParagraphBlock | Heading1Block | Heading2Block | ImageBlockBlock | UnsupportedBlockBlock | VideBlockBlock | LinkPreviewBlockBlock | AudioBlockBlock | FileBlockBlock | PdfBlockBlock | BookmarkBlockBlock | EmbedBlockBlock | TableBlockBlock | TableOfContentsBlockBlock | TableRowBlockBlock | LinkToPageBlockBlock | ColumnBlockBlock | ColumnListBlockBlock | CalloutBlock | DividerBlockBlock | EquationBlock | CodeBlock | ChildDatabaseBlock | ChildPageBlock | SyncedBlockBlock | TemplateBlock | ToggleBlock | ToDoBlock | QuoteBlock | NumberedListItemBlock | BulletedListItemBlock | Heading3Block
}

// block returning interface builder
func GetBlock[BTR AnyBlock](typeOfBlock string) (BTR, error) {
	var calcVal interface{}
	switch typeOfBlock {
	case ImageBlockType:
		calcVal = ImageBlockBlock{}
	case VideBlockType:
		calcVal = VideBlockBlock{}
	case UnsupportedBlockType:
		calcVal = UnsupportedBlockBlock{}
	case LinkPreviewBlockType:
		calcVal = LinkPreviewBlockBlock{}
	case AudioBlockType:
		calcVal = AudioBlockBlock{}
	case FileBlockType:
		calcVal = FileBlockBlock{}
	case PdfBlockType:
		calcVal = PdfBlockBlock{}
	case BookmarkBlockType:
		calcVal = BookmarkBlockBlock{}
	case EmbedBlockType:
		calcVal = EmbedBlockBlock{}
	case TableRowBlockType:
		calcVal = TableRowBlockBlock{}
	case TableBlockType:
		calcVal = TableBlockBlock{}
	case LinkToPageBlockType:
		calcVal = LinkToPageBlockBlock{}
	case ColumnBlockType:
		calcVal = ColumnBlockBlock{}
	case ColumnListBlockType:
		calcVal = ColumnListBlockBlock{}
	case TableOfContentsBlockType:
		calcVal = TableOfContentsBlockBlock{}
	case CalloutType:
		calcVal = CalloutBlock{}
	case DividerBlockType:
		calcVal = DividerBlockBlock{}
	case EquationType:
		calcVal = EquationBlock{}
	case CodeType:
		calcVal = CodeBlock{}
	case ChildDatabaseType:
		calcVal = ChildDatabaseBlock{}
	case ChildPageType:
		calcVal = ChildPageBlock{}
	case SyncedBlockType:
		calcVal = SyncedBlockBlock{}
	case TemplateType:
		calcVal = TemplateBlock{}
	case ParagraphType:
		calcVal = ParagraphBlock{}
	case ToggleType:
		calcVal = ToggleBlock{}
	case ToDoType:
		calcVal = ToDoBlock{}
	case QuoteType:
		calcVal = QuoteBlock{}
	case NumberedListItemType:
		calcVal = NumberedListItemBlock{}
	case BulletedListItemType:
		calcVal = BulletedListItemBlock{}
	case Heading3Type:
		calcVal = Heading3Block{}
	case Heading2Type:
		calcVal = Heading2Block{}
	case Heading1Type:
		calcVal = Heading1Block{}
	default:
		calcVal = UnsupportedBlockBlock{}
	}
	block, ok := calcVal.(BTR)
	if ok {
		return block, nil
	} else {
		return block, errors.New("fudge city")
	}
}
