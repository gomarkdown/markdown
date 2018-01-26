## Tracking perf changes

Initial performance:
```
$ go test -bench=. -test.benchmem
goos: darwin
goarch: amd64
pkg: github.com/gomarkdown/markdown
BenchmarkEscapeHTML-8                           	 2000000	       823 ns/op	       0 B/op	       0 allocs/op
BenchmarkSmartDoubleQuotes-8                    	  300000	      5033 ns/op	    9872 B/op	      56 allocs/op
BenchmarkReferenceAmps-8                        	  100000	     19538 ns/op	   26776 B/op	     150 allocs/op
BenchmarkReferenceAutoLinks-8                   	  100000	     17574 ns/op	   24544 B/op	     132 allocs/op
BenchmarkReferenceBackslashEscapes-8            	   30000	     50977 ns/op	   76752 B/op	     243 allocs/op
BenchmarkReferenceBlockquotesWithCodeBlocks-8   	  200000	      8546 ns/op	   12864 B/op	      65 allocs/op
BenchmarkReferenceCodeBlocks-8                  	  200000	      9000 ns/op	   14912 B/op	      70 allocs/op
BenchmarkReferenceCodeSpans-8                   	  200000	      8856 ns/op	   14992 B/op	      69 allocs/op
BenchmarkReferenceHardWrappedPara-8             	  200000	      6599 ns/op	   11312 B/op	      57 allocs/op
BenchmarkReferenceHorizontalRules-8             	  100000	     15483 ns/op	   23536 B/op	      98 allocs/op
BenchmarkReferenceInlineHTMLAdvances-8          	  200000	      6839 ns/op	   12150 B/op	      62 allocs/op
BenchmarkReferenceInlineHTMLSimple-8            	  100000	     19940 ns/op	   28488 B/op	     117 allocs/op
BenchmarkReferenceInlineHTMLComments-8          	  200000	      7455 ns/op	   13440 B/op	      64 allocs/op
BenchmarkReferenceLinksInline-8                 	  100000	     16425 ns/op	   23664 B/op	     147 allocs/op
BenchmarkReferenceLinksReference-8              	   30000	     54895 ns/op	   66464 B/op	     416 allocs/op
BenchmarkReferenceLinksShortcut-8               	  100000	     17647 ns/op	   23776 B/op	     158 allocs/op
BenchmarkReferenceLiterQuotesInTitles-8         	  200000	      9367 ns/op	   14832 B/op	      95 allocs/op
BenchmarkReferenceMarkdownBasics-8              	   10000	    129772 ns/op	  130848 B/op	     378 allocs/op
BenchmarkReferenceMarkdownSyntax-8              	    3000	    502365 ns/op	  461411 B/op	    1411 allocs/op
BenchmarkReferenceNestedBlockquotes-8           	  200000	      7028 ns/op	   12688 B/op	      64 allocs/op
BenchmarkReferenceOrderedAndUnorderedLists-8    	   20000	     79686 ns/op	  107520 B/op	     374 allocs/op
BenchmarkReferenceStrongAndEm-8                 	  200000	     10020 ns/op	   17792 B/op	      78 allocs/op
BenchmarkReferenceTabs-8                        	  200000	     12025 ns/op	   18224 B/op	      81 allocs/op
BenchmarkReferenceTidyness-8                    	  200000	      8985 ns/op	   14432 B/op	      71 allocs/op
PASS
ok  	github.com/gomarkdown/markdown	45.375s
```

After switching to using interface{} for Node.Data:
```
BenchmarkEscapeHTML-8                           	 2000000	       929 ns/op	       0 B/op	       0 allocs/op
BenchmarkSmartDoubleQuotes-8                    	  300000	      5126 ns/op	    9248 B/op	      56 allocs/op
BenchmarkReferenceAmps-8                        	  100000	     19927 ns/op	   17880 B/op	     154 allocs/op
BenchmarkReferenceAutoLinks-8                   	  100000	     20732 ns/op	   17360 B/op	     141 allocs/op
BenchmarkReferenceBackslashEscapes-8            	   30000	     50267 ns/op	   38128 B/op	     244 allocs/op
BenchmarkReferenceBlockquotesWithCodeBlocks-8   	  200000	      8988 ns/op	   10912 B/op	      67 allocs/op
BenchmarkReferenceCodeBlocks-8                  	  200000	      8611 ns/op	   12256 B/op	      74 allocs/op
BenchmarkReferenceCodeSpans-8                   	  200000	      8256 ns/op	   11248 B/op	      69 allocs/op
BenchmarkReferenceHardWrappedPara-8             	  200000	      6739 ns/op	    9856 B/op	      57 allocs/op
BenchmarkReferenceHorizontalRules-8             	  100000	     15503 ns/op	   15600 B/op	     104 allocs/op
BenchmarkReferenceInlineHTMLAdvances-8          	  200000	      6874 ns/op	   10278 B/op	      62 allocs/op
BenchmarkReferenceInlineHTMLSimple-8            	  100000	     22271 ns/op	   18552 B/op	     121 allocs/op
BenchmarkReferenceInlineHTMLComments-8          	  200000	      8315 ns/op	   10736 B/op	      64 allocs/op
BenchmarkReferenceLinksInline-8                 	  100000	     16155 ns/op	   16912 B/op	     152 allocs/op
BenchmarkReferenceLinksReference-8              	   30000	     52387 ns/op	   38192 B/op	     445 allocs/op
BenchmarkReferenceLinksShortcut-8               	  100000	     17111 ns/op	   16592 B/op	     167 allocs/op
BenchmarkReferenceLiterQuotesInTitles-8         	  200000	      9164 ns/op	   12048 B/op	      97 allocs/op
BenchmarkReferenceMarkdownBasics-8              	   10000	    129262 ns/op	   87264 B/op	     416 allocs/op
BenchmarkReferenceMarkdownSyntax-8              	    3000	    496873 ns/op	  293906 B/op	    1559 allocs/op
BenchmarkReferenceNestedBlockquotes-8           	  200000	      6854 ns/op	   10192 B/op	      64 allocs/op
BenchmarkReferenceOrderedAndUnorderedLists-8    	   20000	     79633 ns/op	   55024 B/op	     447 allocs/op
BenchmarkReferenceStrongAndEm-8                 	  200000	      9637 ns/op	   12176 B/op	      78 allocs/op
BenchmarkReferenceTabs-8                        	  100000	     12164 ns/op	   13776 B/op	      87 allocs/op
BenchmarkReferenceTidyness-8                    	  200000	      8677 ns/op	   11296 B/op	      75 allocs/op
```

Not necessarily faster, but uses less bytes per op (but sometimes more allocs).
