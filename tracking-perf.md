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

After tweaking the API:
```
$ ./s/run-bench.sh

go test -bench=. -test.benchmem
goos: darwin
goarch: amd64
pkg: github.com/gomarkdown/markdown
BenchmarkEscapeHTML-8                           	 2000000	       834 ns/op	       0 B/op	       0 allocs/op
BenchmarkSmartDoubleQuotes-8                    	  300000	      3486 ns/op	    6160 B/op	      27 allocs/op
BenchmarkReferenceAmps-8                        	  100000	     18158 ns/op	   14792 B/op	     125 allocs/op
BenchmarkReferenceAutoLinks-8                   	  100000	     16824 ns/op	   14272 B/op	     112 allocs/op
BenchmarkReferenceBackslashEscapes-8            	   30000	     44066 ns/op	   35040 B/op	     215 allocs/op
BenchmarkReferenceBlockquotesWithCodeBlocks-8   	  200000	      6868 ns/op	    7824 B/op	      38 allocs/op
BenchmarkReferenceCodeBlocks-8                  	  200000	      7157 ns/op	    9168 B/op	      45 allocs/op
BenchmarkReferenceCodeSpans-8                   	  200000	      6663 ns/op	    8160 B/op	      40 allocs/op
BenchmarkReferenceHardWrappedPara-8             	  300000	      4821 ns/op	    6768 B/op	      28 allocs/op
BenchmarkReferenceHorizontalRules-8             	  100000	     13033 ns/op	   12512 B/op	      75 allocs/op
BenchmarkReferenceInlineHTMLAdvances-8          	  300000	      4998 ns/op	    7190 B/op	      33 allocs/op
BenchmarkReferenceInlineHTMLSimple-8            	  100000	     17696 ns/op	   15464 B/op	      92 allocs/op
BenchmarkReferenceInlineHTMLComments-8          	  300000	      5506 ns/op	    7648 B/op	      35 allocs/op
BenchmarkReferenceLinksInline-8                 	  100000	     14450 ns/op	   13824 B/op	     123 allocs/op
BenchmarkReferenceLinksReference-8              	   30000	     52561 ns/op	   35104 B/op	     416 allocs/op
BenchmarkReferenceLinksShortcut-8               	  100000	     15616 ns/op	   13504 B/op	     138 allocs/op
BenchmarkReferenceLiterQuotesInTitles-8         	  200000	      7772 ns/op	    8960 B/op	      68 allocs/op
BenchmarkReferenceMarkdownBasics-8              	   10000	    121436 ns/op	   84176 B/op	     387 allocs/op
BenchmarkReferenceMarkdownSyntax-8              	    3000	    487404 ns/op	  290818 B/op	    1530 allocs/op
BenchmarkReferenceNestedBlockquotes-8           	  300000	      5098 ns/op	    7104 B/op	      35 allocs/op
BenchmarkReferenceOrderedAndUnorderedLists-8    	   20000	     74422 ns/op	   51936 B/op	     418 allocs/op
BenchmarkReferenceStrongAndEm-8                 	  200000	      7888 ns/op	    9088 B/op	      49 allocs/op
BenchmarkReferenceTabs-8                        	  200000	     10061 ns/op	   10688 B/op	      58 allocs/op
BenchmarkReferenceTidyness-8                    	  200000	      7152 ns/op	    8208 B/op	      46 allocs/op
PASS
ok  	github.com/gomarkdown/markdown	40.809s
```
