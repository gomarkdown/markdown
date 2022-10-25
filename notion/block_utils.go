package notion

import (
	"errors"
	"fmt"
)

type AnyBlock interface {
	AudioBlockBlock |
		BookmarkBlockBlock |
		BulletedListItemBlock |
		CalloutBlock |
		ChildDatabaseBlock |
		ChildPageBlock |
		CodeBlock |
		ColumnBlockBlock |
		ColumnListBlockBlock |
		DividerBlockBlock |
		EmbedBlockBlock |
		EquationBlock |
		FileBlockBlock |
		Heading1Block |
		Heading2Block |
		Heading3Block |
		ImageBlockBlock |
		LinkPreviewBlockBlock |
		LinkToPageBlockBlock |
		NumberedListItemBlock |
		ParagraphBlock |
		PdfBlockBlock |
		QuoteBlock |
		SyncedBlockBlock |
		TableBlockBlock |
		TableOfContentsBlockBlock |
		TableRowBlockBlock |
		TemplateBlock |
		ToDoBlock |
		ToggleBlock |
		UnsupportedBlockBlock |
		VideBlockBlock
}

type AnyBlockType interface {
	AudioBlockTypeString |
		BookmarkBlockTypeString |
		BreadcrumbTypeString |
		BulletedListItemTypeString |
		CalloutTypeString |
		ChildDatabaseTypeString |
		ChildPageTypeString |
		CodeTypeString |
		ColumnBlockTypeString |
		ColumnListBlockTypeString |
		DividerBlockTypeString |
		EmbedBlockTypeString |
		EquationTypeString |
		FileBlockTypeString |
		Heading1TypeString |
		Heading2TypeString |
		Heading3TypeString |
		ImageBlockTypeString |
		LinkPreviewBlockTypeString |
		LinkToPageBlockTypeString |
		NumberedListItemTypeString |
		ParagraphTypeString |
		PdfBlockTypeString |
		QuoteTypeString |
		RichTextTypeString |
		SyncedBlockTypeString |
		TableBlockTypeString |
		TableOfContentsBlockTypeString |
		TableRowBlockTypeString |
		TemplateTypeString |
		ToDoTypeString |
		ToggleTypeString |
		UnsupportedBlockTypeString |
		VideBlockTypeString
}

// GetBlock returns a properly instantiated AnyBlock with the type field set to AnyBlockType, or an error
func GetBlock[BTR AnyBlock, BTRT AnyBlockType](typeOfBlock BTRT) (BTR, error) {
	var calcVal interface{}
	actualValue, ok := any(typeOfBlock).(BTRT)
	if !ok {
		fmt.Printf("you sent me this garbo: %s, why?", typeOfBlock)
	}
	aVS := string(actualValue)
	switch aVS {
	case string(ImageBlockType):
		calcVal = ImageBlockBlock{
			Type: ImageBlockType,
		}
	case string(VideBlockType):
		calcVal = VideBlockBlock{
			Type: VideBlockType,
		}
	case string(LinkPreviewBlockType):
		calcVal = LinkPreviewBlockBlock{
			Type: LinkPreviewBlockType,
		}
	case string(AudioBlockType):
		calcVal = AudioBlockBlock{
			Type: AudioBlockType,
		}
	case string(FileBlockType):
		calcVal = FileBlockBlock{
			Type: FileBlockType,
		}
	case string(PdfBlockType):
		calcVal = PdfBlockBlock{
			Type: PdfBlockType,
		}
	case string(BookmarkBlockType):
		calcVal = BookmarkBlockBlock{
			Type: BookmarkBlockType,
		}
	case string(EmbedBlockType):
		calcVal = EmbedBlockBlock{
			Type: EmbedBlockType,
		}
	case string(TableRowBlockType):
		calcVal = TableRowBlockBlock{
			Type: TableRowBlockType,
		}
	case string(TableBlockType):
		calcVal = TableBlockBlock{
			Type: TableBlockType,
		}
	case string(LinkToPageBlockType):
		calcVal = LinkToPageBlockBlock{
			Type: LinkToPageBlockType,
		}
	case string(ColumnBlockType):
		calcVal = ColumnBlockBlock{
			Type: ColumnBlockType,
		}
	case string(ColumnListBlockType):
		calcVal = ColumnListBlockBlock{
			Type: ColumnListBlockType,
		}
	case string(TableOfContentsBlockType):
		calcVal = TableOfContentsBlockBlock{
			Type: TableOfContentsBlockType,
		}
	case string(CalloutType):
		calcVal = CalloutBlock{
			Type: CalloutType,
		}
	case string(DividerBlockType):
		calcVal = DividerBlockBlock{
			Type: DividerBlockType,
		}
	case string(EquationType):
		calcVal = EquationBlock{
			Type: EquationType,
		}
	case string(CodeType):
		calcVal = CodeBlock{
			Type: CodeType,
		}
	case string(ChildDatabaseType):
		calcVal = ChildDatabaseBlock{
			Type: ChildDatabaseType,
		}
	case string(ChildPageType):
		calcVal = ChildPageBlock{
			Type: ChildPageType,
		}
	case string(SyncedBlockType):
		calcVal = SyncedBlockBlock{
			Type: SyncedBlockType,
		}
	case string(TemplateType):
		calcVal = TemplateBlock{
			Type: TemplateType,
		}
	case string(ParagraphType):
		calcVal = ParagraphBlock{
			Type: ParagraphType,
		}
	case string(ToggleType):
		calcVal = ToggleBlock{
			Type: ToggleType,
		}
	case string(ToDoType):
		calcVal = ToDoBlock{
			Type: ToDoType,
		}
	case string(QuoteType):
		calcVal = QuoteBlock{
			Type: QuoteType,
		}
	case string(NumberedListItemType):
		calcVal = NumberedListItemBlock{
			Type: NumberedListItemType,
		}
	case string(BulletedListItemType):
		calcVal = BulletedListItemBlock{
			Type: BulletedListItemType,
		}
	case string(Heading3Type):
		calcVal = Heading3Block{
			Type: Heading3Type,
		}
	case string(Heading2Type):
		calcVal = Heading2Block{
			Type: Heading2Type,
		}
	case string(Heading1Type):
		calcVal = Heading1Block{
			Type: Heading1Type,
		}
	default:
		calcVal = UnsupportedBlockBlock{
			Type: "unsupported",
		}
	}
	block, okTwo := calcVal.(BTR)
	if ok {
		return block, errors.New("could not even begin to parse input string")
	} else if okTwo {
		return block, nil
	} else {
		return block, errors.New("we couldn't coerce your block into a valid one")
	}
}
