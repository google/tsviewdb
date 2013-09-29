#!/bin/sh
# TSViewDB time-series data visualization server resource build script.

set -e  # Die if any command fails.

###############################################################################
# SETUP
###############################################################################

if [ -z $GOPATH ]; then
  echo "FAILED: Must set GOPATH environment variable."
  exit
fi

TSVIEWDBROOT=${TSVIEWDBROOT:=$GOPATH/src/github.com/google/tsviewdb}

mkdir -p $TSVIEWDBROOT/resources
TOOLSDIR=$TSVIEWDBROOT/third_party_tools

mkdir -p $TSVIEWDBROOT/resources/fonts
cp $GOPATH/src/code.google.com/p/plotinum/vg/fonts/* $TSVIEWDBROOT/resources/fonts

cd $TSVIEWDBROOT/resources_src

###############################################################################
# JavaScript Compilation
###############################################################################

# Note: Compiled Dygraph JavaScript here is for inline-graph handler.
echo "Compressing Dygraph JavaScript..."
java -jar $TOOLSDIR/compiler.jar \
  --compilation_level SIMPLE_OPTIMIZATIONS \
  --externs dygraph_externs.js \
  --js third_party/dygraph-combined_1.0.0.js \
  --js_output_file $TSVIEWDBROOT/resources/dygraph-compiled.js

# Note: Module compiled Dygraph JavaScript here is for TSViewDB UI handler.
echo 'Compressing main JavaScript...'
java -jar $TOOLSDIR/compiler.jar \
  --compilation_level SIMPLE_OPTIMIZATIONS \
  --externs third_party/jquery-1.9.externs.js \
  --externs dygraph_externs.js \
  --js third_party/dygraph-combined_1.0.0.js \
  --js third_party/goog/base.js \
  --js third_party/goog/crypt/hash.js \
  --js third_party/goog/crypt/md5.js \
  --js third_party/jquery_bbq/jquery.ba-bbq.js \
  --js third_party/jquery_ui_dragslider/dragslider.js \
  --js types.js \
  --js constants.js \
  --js params.js \
  --js histogram.js \
  --js graphs.js \
  --js quickdates.js \
  --js embed.js \
  --js aggsliders.js \
  --js scales.js \
  --js query_change.js \
  --js agg_resize.js \
  --js mbuttons.js \
  --js config.js \
  --js tsviewdb.js \
  --module head:19 \
  --module tail:1:head \
  --module_output_path_prefix $TSVIEWDBROOT/resources/

echo 'Compressing search box JavaScript...'
java -jar $TOOLSDIR/compiler.jar \
  --compilation_level SIMPLE_OPTIMIZATIONS \
  --externs third_party/jquery-1.9.externs.js \
  --js search_box_handler.js \
  --js_output_file $TSVIEWDBROOT/resources/search_box_handler-compiled.js


###############################################################################
# CSS Minification
###############################################################################

# Note: CSS has been run through the csstidy program with these flags:
#
# csstidy tsviewdb.css\
#   --compress_colors=false\
#   --compress_font-weight=false\
#   tsviewdb.tidy.css
#
# ... followed by hand edits to:
# (1) Fix the indentation.
# (2) Add a space after each property ':'
# (3) Put multiple selectors on different lines.
# (4) Add the comments back in.
#

echo 'Compressing CSS...'
java -jar $TOOLSDIR/yuicompressor-2.4.8.jar tsviewdb.css > $TSVIEWDBROOT/resources/tsviewdb.css


###############################################################################
# HTML Minification
###############################################################################

# Note: All HTML source has been run through the tidy program with this config:
#  output-html: yes
#  indent: auto
#  indent-spaces: 2
#  show-warnings: yes
#  wrap: 200
#  hide-endtags: yes
#  tidy-mark: no
#
# ... followed by hand edits to:
# (1) Put the script elements on one-line.
# (2) Omit the type attribute for stylesheets and scripts.
#

echo 'Compressing HTML...'
java -jar $TOOLSDIR/htmlcompressor-1.5.3.jar -o $TSVIEWDBROOT/resources/ --type html .

###############################################################################
# Export Static
###############################################################################

echo 'Exporting to resources directory...'
cp $TSVIEWDBROOT/resources_static/* $TSVIEWDBROOT/resources

echo '... all done.'
