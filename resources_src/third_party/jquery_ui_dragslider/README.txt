Example Use:

HTML:

  <head>
    <script type="text/javascript" src="https://ajax.googleapis.com/ajax/libs/jquery/1.6.1/jquery.min.js"></script>
    <script type="text/javascript" src="https://ajax.googleapis.com/ajax/libs/jqueryui/1.8.13/jquery-ui.min.js"></script>
    <script type="text/javascript" src="http://pastebin.com/raw.php?i=8ugkXxth"></script>
    <link rel="stylesheet" type="text/css" href="http://ajax.googleapis.com/ajax/libs/jqueryui/1.8.10/themes/ui-lightness/jquery-ui.css" />
  </head>
  <body>
    <div id="slider" style=""></div>drag me in the middle!
  </body>

Javascript:

  $(function(){
    $('#slider').dragslider({
    animate: true,
    range: true,
    rangeDrag: true,
    values: [30, 70]
    });
  });
