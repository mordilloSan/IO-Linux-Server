import{r as i,ac as j,j as e,G as c,T as r,p as l,J as d,D as u,q as h,P as p}from"./index-DbTWBLWr.js";const y=l(p)(h),v=l(u)(h),m=s=>{if(s<1024)return`${s} B`;const t=s/1024;return t<1024?`${t.toFixed(1)} KB`:`${(t/1024).toFixed(1)} MB`},g=()=>{const{socket:s}=i.useContext(j),[t,x]=i.useState(null);return i.useEffect(()=>{if(!s)return;const n=a=>{try{const o=JSON.parse(a.data);x(o)}catch(o){console.error("❌ Failed to parse WebSocket data:",o)}};return s.addEventListener("message",n),()=>{s.removeEventListener("message",n)}},[s]),e.jsxs(c,{container:!0,spacing:6,children:[e.jsxs(c,{size:12,children:[e.jsx(r,{variant:"h3",gutterBottom:!0,children:"Live System Stats"}),e.jsx(v,{my:4})]}),e.jsx(c,{size:12,children:e.jsx(y,{p:4,children:t?e.jsxs(e.Fragment,{children:[e.jsx(r,{variant:"h6",children:"🧠 CPU"}),e.jsx(d,{mb:2,children:t.cpu.map((n,a)=>e.jsxs(r,{children:["Core ",a+1,": ",n.toFixed(1),"%"]},a))}),e.jsx(r,{variant:"h6",children:"💾 Memory"}),e.jsx(d,{mb:2,children:e.jsxs(r,{children:[t.memory.usedPercent.toFixed(1),"% used (",(t.memory.used/1024/1024/1024).toFixed(1)," GB of"," ",(t.memory.total/1024/1024/1024).toFixed(1)," GB)"]})}),e.jsx(r,{variant:"h6",children:"🌐 Network"}),e.jsx(d,{children:Object.entries(t.network).map(([n,a])=>e.jsxs(r,{children:[n,": ↑ ",m(a.bytesSent)," ↓"," ",m(a.bytesRecv)]},n))})]}):e.jsx(r,{children:"Loading system stats..."})})})]})};export{g as default};
