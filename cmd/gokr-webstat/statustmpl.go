package main

const statusTmpl = `<!DOCTYPE html>
<html>
  <head>
    <title>{{ .Hostname }} - webstat</title>
<style type="text/css">
body {
  background-color: #eee;
}
#readings {
  background-color: black;
  border: 1px solid grey;
  font-family: monospace;
}
th {
  color: white;
  text-align: left;
}
</style>
  </head>
  <body>
<h1>{{ .Hostname }} - webstat</h1>
<table id="readings">
<thead>
<tr>
{{ range $idx, $val := .Headers }}
<th>{{ $val }}</th>
{{ end }}
</tr>
</thead>
<tbody>
<tr><td>&nbsp;</td></tr>
<tr><td>&nbsp;</td></tr>
<tr><td>&nbsp;</td></tr>
<tr><td>&nbsp;</td></tr>
<tr><td>&nbsp;</td></tr>
<tr><td>&nbsp;</td></tr>
<tr><td>&nbsp;</td></tr>
<tr><td>&nbsp;</td></tr>
<tr><td>&nbsp;</td></tr>
<tr><td>&nbsp;</td></tr>
<tr><td>&nbsp;</td></tr>
<tr><td>&nbsp;</td></tr>
<tr><td>&nbsp;</td></tr>
<tr><td>&nbsp;</td></tr>
<tr><td>&nbsp;</td></tr>
<tr><td>&nbsp;</td></tr>
<tr><td>&nbsp;</td></tr>
<tr><td>&nbsp;</td></tr>
<tr><td>&nbsp;</td></tr>
<tr><td>&nbsp;</td></tr>
</tbody>
</table>
<script type="text/javascript">
var readingstbody = document.getElementById('readings').children[1];
var readings = new EventSource('/readings');
readings.onmessage = function(e) {
  var parts = JSON.parse(e.data);
  var innerHTML = '';

  for (part of parts) {
    innerHTML += part;
  }
  for (var i = 0; i < 19; i++) {
    // copy from i+1 to i, overwriting the oldest entry
    readingstbody.children[i].innerHTML = readingstbody.children[i+1].innerHTML
  }
  readingstbody.children[19].innerHTML = innerHTML;
}
</script>
  </body>
</html>`
