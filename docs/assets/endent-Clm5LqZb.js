import{r as b,g as j}from"./dedent-DLT5raRU.js";import{r as D}from"./objectorarray-BuWbUJTR.js";import{r as q}from"./fast-json-parse-CjGQNN3V.js";var l={},g;function E(){if(g)return l;g=1;var o=l&&l.__importDefault||function(t){return t&&t.__esModule?t:{default:t}};Object.defineProperty(l,"__esModule",{value:!0});const h=o(b()),m=o(D()),p=o(q()),i="twhZNwxI1aFG3r4";function c(t,...d){let r="";for(let a=0;a<t.length;a++)if(r+=t[a],a<d.length){let e=d[a],f=!1;if(p.default(e).value&&(e=p.default(e).value,f=!0),e&&e[i]||f){let u=r.split(`
`),s=u[u.length-1].search(/\S/),n=s>0?" ".repeat(s):"";(f?JSON.stringify(e,null,2):e[i]).split(`
`).forEach((v,y)=>{y>0?r+=`
`+n+v:r+=v})}else if(typeof e=="string"&&e.includes(`
`)){let u=r.match(/(?:^|\n)( *)$/);if(typeof e=="string"){let s=u?u[1]:"";r+=e.split(`
`).map((n,_)=>(n=i+n,_===0?n:`${s}${n}`)).join(`
`)}else r+=e}else r+=e}return r=h.default(r),r.split(i).join("")}return c.pretty=t=>m.default(t)?{[i]:JSON.stringify(t,null,2)}:t,l.default=c,l}var J=E();const w=j(J);export{w as e};
