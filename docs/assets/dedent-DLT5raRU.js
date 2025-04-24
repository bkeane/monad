function d(e){return e&&e.__esModule&&Object.prototype.hasOwnProperty.call(e,"default")?e.default:e}var o={exports:{}},p;function g(){return p||(p=1,function(e){function c(l){var u=void 0;typeof l=="string"?u=[l]:u=l.raw;for(var r="",t=0;t<u.length;t++)r+=u[t].replace(/\\\n[ \t]*/g,"").replace(/\\`/g,"`"),t<(arguments.length<=1?0:arguments.length-1)&&(r+=arguments.length<=t+1?void 0:arguments[t+1]);var f=r.split(`
`),n=null;return f.forEach(function(a){var s=a.match(/^(\s+)\S+/);if(s){var i=s[1].length;n?n=Math.min(n,i):n=i}}),n!==null&&(r=f.map(function(a){return a[0]===" "?a.slice(n):a}).join(`
`)),r=r.trim(),r.replace(/\\n/g,`
`)}e.exports=c}(o)),o.exports}export{d as g,g as r};
