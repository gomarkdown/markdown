package notion

import (
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

type RichTextTypeString string

const RichTextType RichTextTypeString = "rich_text"

type RichText struct {
	BlockBase
	Type        RichTextTypeString  `json:"type"`
	ID          BlockID             `json:"id,omitempty"`
	Text        RichTextTextContent `json:"rich_text"`
	PlaintText  string              `json:"plain_text"`
	Annotations AnnotationResponse  `json:"annotations"`
	HREF        string              `json:"href"`
}

func (rt RichText) TypeInternal() string {
	return string(RichTextType)
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

type ParagraphTypeString string

const ParagraphType ParagraphTypeString = "paragraph"

type ParagraphBlock struct {
	BlockBase
	Type      ParagraphTypeString `json:"type"`
	ID        BlockID             `json:"id,omitempty"`
	Paragraph Paragraph           `json:"paragraph"`
	Parent    BlockHolster        `json:"parent"`
	Object    BlockType           `json:"object"`
}

func (pb ParagraphBlock) TypeInternal() string {
	return string(ParagraphType)
}
func (pb ParagraphBlock) ActualBlock() interface{} {
	return pb
}

// parent block stubs
type DataBaseBlock struct {
	BlockBase
	Type     string     `json:"type"`
	ID       BlockID    `json:"id,omitempty"`
	DataBase DataBaseID `json:"database_id"`
}

type PageBlock struct {
	BlockBase
	Type   string  `json:"type"`
	ID     BlockID `json:"id,omitempty"`
	PageID PageID  `json:"page_id"`
}

type BlockBlock struct {
	BlockBase
	Type    string  `json:"type"`
	ID      BlockID `json:"id,omitempty"`
	BlockID BlockID `json:"block_id"`
}

type WorkspaceBlock struct {
	BlockBase
	Type      string    `json:"type"`
	ID        BlockID   `json:"id,omitempty"`
	WorkSpace WorkSpace `json:"workspace"`
}

// all supported blocks
type Heading1 struct {
	RichText []RichText `json:"rich_text"`
	Color    ApiColor   `json:"color"`
}

type Heading1TypeString string

const Heading1Type Heading1TypeString = "heading_1"

type Heading1Block struct {
	BlockBase
	Type     Heading1TypeString `json:"type"`
	ID       BlockID            `json:"id,omitempty"`
	Heading1 Heading1           `json:"heading_1"`
	Parent   BlockHolster       `json:"parent"`
	Object   BlockType          `json:"object"`
}

func (bI Heading1Block) TypeInternal() string {
	return string(Heading1Type)
}
func (bI Heading1Block) ActualBlock() interface{} {
	return bI
}

type Heading2 struct {
	RichText []RichText `json:"rich_text"`
	Color    ApiColor   `json:"color"`
}

type Heading2TypeString string

const Heading2Type Heading2TypeString = "heading_2"

type Heading2Block struct {
	BlockBase
	Type     Heading2TypeString `json:"type"`
	ID       BlockID            `json:"id,omitempty"`
	Heading2 Heading2           `json:"heading_2"`
	Parent   BlockHolster       `json:"parent"`
	Object   BlockType          `json:"object"`
}

func (bI Heading2Block) TypeInternal() string {
	return string(Heading2Type)
}
func (bI Heading2Block) ActualBlock() interface{} {
	return bI
}

type Heading3 struct {
	RichText []RichText `json:"rich_text"`
	Color    ApiColor   `json:"color"`
}

type Heading3TypeString string

const Heading3Type Heading3TypeString = "heading_3"

type Heading3Block struct {
	BlockBase
	Type     Heading3TypeString `json:"type"`
	ID       BlockID            `json:"id,omitempty"`
	Heading3 Heading3           `json:"heading_3"`
	Parent   BlockHolster       `json:"parent"`
	Object   BlockType          `json:"object"`
}

func (bI Heading3Block) TypeInternal() string {
	return string(Heading3Type)
}
func (bI Heading3Block) ActualBlock() interface{} {
	return bI
}

type BulletedListItem struct {
	RichText []RichText `json:"rich_text"`
	Color    ApiColor   `json:"color"`
}

type BulletedListItemTypeString string

const BulletedListItemType BulletedListItemTypeString = "bulleted_list_item"

type BulletedListItemBlock struct {
	BlockBase
	Type             BulletedListItemTypeString `json:"type"`
	ID               BlockID                    `json:"id,omitempty"`
	BulletedListItem BulletedListItem           `json:"bulleted_list_item"`
	Parent           BlockHolster               `json:"parent"`
	Object           BlockType                  `json:"object"`
}

func (bI BulletedListItemBlock) TypeInternal() string {
	return string(BulletedListItemType)
}
func (bI BulletedListItemBlock) ActualBlock() interface{} {
	return bI
}

type NumberedListItem struct {
	RichText []RichText `json:"rich_text"`
	Color    ApiColor   `json:"color"`
}

type NumberedListItemTypeString string

const NumberedListItemType NumberedListItemTypeString = "numbered_list_item"

type NumberedListItemBlock struct {
	BlockBase
	Type             NumberedListItemTypeString `json:"type"`
	ID               BlockID                    `json:"id,omitempty"`
	NumberedListItem NumberedListItem           `json:"numbered_list_item"`
	Parent           BlockHolster               `json:"parent"`
	Object           BlockType                  `json:"object"`
}

func (bI NumberedListItemBlock) TypeInternal() string {
	return string(NumberedListItemType)
}
func (bI NumberedListItemBlock) ActualBlock() interface{} {
	return bI
}

type Quote struct {
	RichText []RichText `json:"rich_text"`
	Color    ApiColor   `json:"color"`
}

type QuoteTypeString string

const QuoteType QuoteTypeString = "quote"

type QuoteBlock struct {
	BlockBase
	Type   QuoteTypeString `json:"type"`
	ID     BlockID         `json:"id,omitempty"`
	Quote  Quote           `json:"quote"`
	Parent BlockHolster    `json:"parent"`
	Object BlockType       `json:"object"`
}

func (bI QuoteBlock) TypeInternal() string {
	return string(QuoteType)
}
func (bI QuoteBlock) ActualBlock() interface{} {
	return bI
}

type ToDo struct {
	RichText []RichText `json:"rich_text"`
	Color    ApiColor   `json:"color"`
	Checked  bool       `json:"checked"`
}

type ToDoTypeString string

const ToDoType ToDoTypeString = "to_do"

type ToDoBlock struct {
	BlockBase
	Type   ToDoTypeString `json:"type"`
	ID     BlockID        `json:"id,omitempty"`
	ToDo   ToDo           `json:"to_do"`
	Parent BlockHolster   `json:"parent"`
	Object BlockType      `json:"object"`
}

func (bI ToDoBlock) TypeInternal() string {
	return string(ToDoType)
}
func (bI ToDoBlock) ActualBlock() interface{} {
	return bI
}

type Toggle struct {
	RichText []RichText `json:"rich_text"`
	Color    ApiColor   `json:"color"`
}

type ToggleTypeString string

const ToggleType ToggleTypeString = "toggle"

type ToggleBlock struct {
	BlockBase
	Type   ToggleTypeString `json:"type"`
	ID     BlockID          `json:"id,omitempty"`
	Toggle Toggle           `json:"toggle"`
	Parent BlockHolster     `json:"parent"`
	Object BlockType        `json:"object"`
}

func (bI ToggleBlock) TypeInternal() string {
	return string(ToggleType)
}
func (bI ToggleBlock) ActualBlock() interface{} {
	return bI
}

type Template struct {
	RichText []RichText `json:"rich_text"`
}

type TemplateTypeString string

const TemplateType TemplateTypeString = "template"

type TemplateBlock struct {
	BlockBase
	Type     TemplateTypeString `json:"type"`
	ID       BlockID            `json:"id,omitempty"`
	Template Template           `json:"template"`
	Parent   BlockHolster       `json:"parent"`
	Object   BlockType          `json:"object"`
}

func (bI TemplateBlock) TypeInternal() string {
	return string(TemplateType)
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

type SyncedBlockTypeString string

const SyncedBlockType SyncedBlockTypeString = "synced_block"

type SyncedBlockBlock struct {
	BlockBase
	Type        SyncedBlockTypeString `json:"type"`
	ID          BlockID               `json:"id,omitempty"`
	SyncedBlock SyncedBlock           `json:"synced_block"`
	Parent      BlockHolster          `json:"parent"`
	Object      BlockType             `json:"object"`
}

func (bI SyncedBlockBlock) TypeInternal() string {
	return string(SyncedBlockType)
}
func (bI SyncedBlockBlock) ActualBlock() interface{} {
	return bI
}

type ChildPage struct {
	Title string `json:"title"`
}

type ChildPageTypeString string

const ChildPageType ChildPageTypeString = "child_page"

type ChildPageBlock struct {
	BlockBase
	Type      ChildPageTypeString `json:"type"`
	ID        BlockID             `json:"id,omitempty"`
	ChildPage ChildPage           `json:"child_page"`
	Parent    BlockHolster        `json:"parent"`
	Object    BlockType           `json:"object"`
}

func (bI ChildPageBlock) TypeInternal() string {
	return string(ChildPageType)
}
func (bI ChildPageBlock) ActualBlock() interface{} {
	return bI
}

type ChildDatabase struct {
	Title string `json:"title"`
}

type ChildDatabaseTypeString string

const ChildDatabaseType ChildDatabaseTypeString = "child_database"

type ChildDatabaseBlock struct {
	BlockBase
	Type          ChildDatabaseTypeString `json:"type"`
	ID            BlockID                 `json:"id,omitempty"`
	ChildDatabase ChildDatabase           `json:"child_database"`
	Parent        BlockHolster            `json:"parent"`
	Object        BlockType               `json:"object"`
}

func (bI ChildDatabaseBlock) TypeInternal() string {
	return string(ChildDatabaseType)
}
func (bI ChildDatabaseBlock) ActualBlock() interface{} {
	return bI
}

type Equation struct {
	Expression string `json:"expression"`
}

type EquationTypeString string

const EquationType EquationTypeString = "equation"

type EquationBlock struct {
	BlockBase
	Type     EquationTypeString `json:"type"`
	ID       BlockID            `json:"id,omitempty"`
	Equation Equation           `json:"equation"`
	Parent   BlockHolster       `json:"parent"`
	Object   BlockType          `json:"object"`
}

func (bI EquationBlock) TypeInternal() string {
	return string(EquationType)
}
func (bI EquationBlock) ActualBlock() interface{} {
	return bI
}

type Code struct {
	RichText []RichText `json:"rich_text"`
	Color    ApiColor   `json:"color"`
	Language string     `json:"language"`
}

type CodeTypeString string

const CodeType CodeTypeString = "code"

type CodeBlock struct {
	BlockBase
	Type   CodeTypeString `json:"type"`
	ID     BlockID        `json:"id,omitempty"`
	Code   Code           `json:"code"`
	Parent BlockHolster   `json:"parent"`
	Object BlockType      `json:"object"`
}

func (bI CodeBlock) TypeInternal() string {
	return string(CodeType)
}
func (bI CodeBlock) ActualBlock() interface{} {
	return bI
}

type Callout struct {
	RichText []RichText  `json:"rich_text"`
	Color    ApiColor    `json:"color"`
	Icon     interface{} `json:"icon"`
}

type CalloutTypeString string

const CalloutType CalloutTypeString = "callout"

type CalloutBlock struct {
	BlockBase
	Type    CalloutTypeString `json:"type"`
	ID      BlockID           `json:"id,omitempty"`
	Callout Callout           `json:"callout"`
	Parent  BlockHolster      `json:"parent"`
	Object  BlockType         `json:"object"`
}

func (bI CalloutBlock) TypeInternal() string {
	return string(CalloutType)
}
func (bI CalloutBlock) ActualBlock() interface{} {
	return bI
}

type DividerBlock struct {
}

type DividerBlockTypeString string

const DividerBlockType DividerBlockTypeString = "divider"

type DividerBlockBlock struct {
	BlockBase
	Type         DividerBlockTypeString `json:"type"`
	ID           BlockID                `json:"id,omitempty"`
	DividerBlock DividerBlock           `json:"divider"`
	Parent       BlockHolster           `json:"parent"`
	Object       BlockType              `json:"object"`
}

func (bI DividerBlockBlock) TypeInternal() string {
	return string(DividerBlockType)
}
func (bI DividerBlockBlock) ActualBlock() interface{} {
	return bI
}

type Breadcrumb struct {
}

type BreadcrumbTypeString string

const BreadcrumbType BreadcrumbTypeString = "breadcrumb"

type BreadcrumbBlock struct {
	BlockBase
	Type       BreadcrumbTypeString `json:"type"`
	ID         BlockID              `json:"id,omitempty"`
	Breadcrumb Breadcrumb           `json:"breadcrumb"`
	Parent     BlockHolster         `json:"parent"`
	Object     BlockType            `json:"object"`
}

func (pb BreadcrumbBlock) TypeInternal() string {
	return string(BreadcrumbType)
}
func (pb BreadcrumbBlock) ActualBlock() interface{} {
	return pb
}

type TableOfContentsBlock struct {
	Color ApiColor `json:"color"`
}

type TableOfContentsBlockTypeString string

const TableOfContentsBlockType TableOfContentsBlockTypeString = "table_of_contents"

type TableOfContentsBlockBlock struct {
	BlockBase
	Type                 TableOfContentsBlockTypeString `json:"type"`
	ID                   BlockID                        `json:"id,omitempty"`
	TableOfContentsBlock TableOfContentsBlock           `json:"table_of_contents"`
	Parent               BlockHolster                   `json:"parent"`
	Object               BlockType                      `json:"object"`
}

func (pb TableOfContentsBlockBlock) TypeInternal() string {
	return string(TableOfContentsBlockType)
}
func (pb TableOfContentsBlockBlock) ActualBlock() interface{} {
	return pb
}

type ColumnListBlock struct {
}

type ColumnListBlockTypeString string

const ColumnListBlockType ColumnListBlockTypeString = "column_list"

type ColumnListBlockBlock struct {
	BlockBase
	Type            ColumnListBlockTypeString `json:"type"`
	ID              BlockID                   `json:"id,omitempty"`
	ColumnListBlock ColumnListBlock           `json:"column_list"`
	Parent          BlockHolster              `json:"parent"`
	Object          BlockType                 `json:"object"`
}

func (pb ColumnListBlockBlock) TypeInternal() string {
	return string(ColumnListBlockType)
}
func (pb ColumnListBlockBlock) ActualBlock() interface{} {
	return pb
}

type ColumnBlock struct {
}

type ColumnBlockTypeString string

const ColumnBlockType ColumnBlockTypeString = "column"

type ColumnBlockBlock struct {
	BlockBase
	Type        ColumnBlockTypeString `json:"type"`
	ID          BlockID               `json:"id,omitempty"`
	ColumnBlock ColumnBlock           `json:"column"`
	Parent      BlockHolster          `json:"parent"`
	Object      BlockType             `json:"object"`
}

func (pb ColumnBlockBlock) TypeInternal() string {
	return string(ColumnBlockType)
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

type LinkToPageBlockTypeString string

const LinkToPageBlockType LinkToPageBlockTypeString = "link_to_page"

type LinkToPageBlockBlock struct {
	BlockBase
	Type            LinkToPageBlockTypeString `json:"type"`
	ID              BlockID                   `json:"id,omitempty"`
	LinkToPageBlock LinkToPageBlock           `json:"link_to_page"`
	Parent          BlockHolster              `json:"parent"`
	Object          BlockType                 `json:"object"`
}

func (pb LinkToPageBlockBlock) TypeInternal() string {
	return string(LinkToPageBlockType)
}
func (pb LinkToPageBlockBlock) ActualBlock() interface{} {
	return pb
}

type TableBlock struct {
	HasColumnHeader bool
	HasRowHeader    bool
	TableWidth      bool
}

type TableBlockTypeString string

const TableBlockType TableBlockTypeString = "table"

type TableBlockBlock struct {
	BlockBase
	Type       TableBlockTypeString `json:"type"`
	ID         BlockID              `json:"id,omitempty"`
	TableBlock TableBlock           `json:"table"`
	Parent     BlockHolster         `json:"parent"`
	Object     BlockType            `json:"object"`
}

func (pb TableBlockBlock) TypeInternal() string {
	return string(TableBlockType)
}
func (pb TableBlockBlock) ActualBlock() interface{} {
	return pb
}

type TableRowBlock struct {
	Cells [][]RichText
}

type TableRowBlockTypeString string

const TableRowBlockType TableRowBlockTypeString = "table_row"

type TableRowBlockBlock struct {
	BlockBase
	Type          TableRowBlockTypeString `json:"type"`
	ID            BlockID                 `json:"id,omitempty"`
	TableRowBlock TableRowBlock           `json:"table_row"`
	Parent        BlockHolster            `json:"parent"`
	Object        BlockType               `json:"object"`
}

func (pb TableRowBlockBlock) TypeInternal() string {
	return string(TableRowBlockType)
}
func (pb TableRowBlockBlock) ActualBlock() interface{} {
	return pb
}

type EmbedBlock struct {
	URL     string
	Caption []RichText
}

type EmbedBlockTypeString string

const EmbedBlockType EmbedBlockTypeString = "embed"

type EmbedBlockBlock struct {
	BlockBase
	Type       EmbedBlockTypeString `json:"type"`
	ID         BlockID              `json:"id,omitempty"`
	EmbedBlock EmbedBlock           `json:"embed"`
	Parent     BlockHolster         `json:"parent"`
	Object     BlockType            `json:"object"`
}

func (pb EmbedBlockBlock) TypeInternal() string {
	return string(EmbedBlockType)
}
func (pb EmbedBlockBlock) ActualBlock() interface{} {
	return pb
}

type BookmarkBlock struct {
	URL     string
	Caption []RichText
}

type BookmarkBlockTypeString string

const BookmarkBlockType BookmarkBlockTypeString = "bookmark"

type BookmarkBlockBlock struct {
	BlockBase
	Type          BookmarkBlockTypeString `json:"type"`
	ID            BlockID                 `json:"id,omitempty"`
	BookmarkBlock BookmarkBlock           `json:"bookmark"`
	Parent        BlockHolster            `json:"parent"`
	Object        BlockType               `json:"object"`
}

func (pb BookmarkBlockBlock) TypeInternal() string {
	return string(BookmarkBlockType)
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

type ImageBlockTypeString string

const ImageBlockType ImageBlockTypeString = "image"

type ImageBlockBlock struct {
	BlockBase
	Type       ImageBlockTypeString `json:"type"`
	ID         BlockID              `json:"id,omitempty"`
	ImageBlock ImageBlock           `json:"image"`
	Parent     BlockHolster         `json:"parent"`
	Object     BlockType            `json:"object"`
}

func (pb ImageBlockBlock) TypeInternal() string {
	return string(ImageBlockType)
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

type VideBlockTypeString string

const VideBlockType VideBlockTypeString = "video"

type VideBlockBlock struct {
	BlockBase
	Type      VideBlockTypeString `json:"type"`
	ID        BlockID             `json:"id,omitempty"`
	VideBlock VideBlock           `json:"video"`
	Parent    BlockHolster        `json:"parent"`
	Object    BlockType           `json:"object"`
}

func (pb VideBlockBlock) TypeInternal() string {
	return string(VideBlockType)
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

type PdfBlockTypeString string

const PdfBlockType PdfBlockTypeString = "pdf"

type PdfBlockBlock struct {
	BlockBase
	Type     PdfBlockTypeString `json:"type"`
	ID       BlockID            `json:"id,omitempty"`
	PdfBlock PdfBlock           `json:"pdf"`
	Parent   BlockHolster       `json:"parent"`
	Object   BlockType          `json:"object"`
}

func (pb PdfBlockBlock) TypeInternal() string {
	return string(PdfBlockType)
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

type FileBlockTypeString string

const FileBlockType FileBlockTypeString = "file"

type FileBlockBlock struct {
	BlockBase
	Type      FileBlockTypeString `json:"type"`
	ID        BlockID             `json:"id,omitempty"`
	FileBlock FileBlock           `json:"file"`
	Parent    BlockHolster        `json:"parent"`
	Object    BlockType           `json:"object"`
}

func (pb FileBlockBlock) TypeInternal() string {
	return string(FileBlockType)
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

type AudioBlockTypeString string

const AudioBlockType AudioBlockTypeString = "audio"

type AudioBlockBlock struct {
	BlockBase
	Type       AudioBlockTypeString `json:"type"`
	ID         BlockID              `json:"id,omitempty"`
	AudioBlock AudioBlock           `json:"audio"`
	Parent     BlockHolster         `json:"parent"`
	Object     BlockType            `json:"object"`
}

func (pb AudioBlockBlock) TypeInternal() string {
	return string(AudioBlockType)
}
func (pb AudioBlockBlock) ActualBlock() interface{} {
	return pb
}

type LinkPreviewBlock struct {
	URL string
}

type LinkPreviewBlockTypeString string

const LinkPreviewBlockType LinkPreviewBlockTypeString = "link_preview"

type LinkPreviewBlockBlock struct {
	BlockBase
	Type             LinkPreviewBlockTypeString `json:"type"`
	ID               BlockID                    `json:"id,omitempty"`
	LinkPreviewBlock LinkPreviewBlock           `json:"link_preview"`
	Parent           BlockHolster               `json:"parent"`
	Object           BlockType                  `json:"object"`
}

func (pb LinkPreviewBlockBlock) TypeInternal() string {
	return string(LinkPreviewBlockType)
}
func (pb LinkPreviewBlockBlock) ActualBlock() interface{} {
	return pb
}

type UnsupportedBlock struct {
}

type UnsupportedBlockTypeString string

const UnsupportedBlockType UnsupportedBlockTypeString = "unsupported"

type UnsupportedBlockBlock struct {
	Type string                     `json:"type"`
	ID   UnsupportedBlockTypeString `json:"id,omitempty"`
	BlockBase
	UnsupportedBlock UnsupportedBlock `json:"unsupported"`
	Parent           BlockHolster     `json:"parent"`
	Object           BlockType        `json:"object"`
}

func (pb UnsupportedBlockBlock) TypeInternal() string {
	return string(UnsupportedBlockType)
}
func (pb UnsupportedBlockBlock) ActualBlock() interface{} {
	return pb
}
