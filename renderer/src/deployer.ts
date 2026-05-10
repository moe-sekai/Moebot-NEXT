// 渲染端"部署者昵称"的全局状态。Go 主程序在启动时以及创建/修改账号后通过
// POST /config 把昵称同步过来；BaseCard 渲染时读取它，把卡片底部 footer 渲染为
// "Moebot NEXT (deployed by <昵称>)"。空昵称时回退为 "Moebot NEXT"。
//
// 进程级状态而非 React Context：Satori 渲染走非 React 树形 layout，没有上下文
// 注入路径，且 Bun 渲染服务是单进程长期常驻，模块作用域足够。

let globalDeployer = "";

export function setDeployer(nickname: string): void {
	globalDeployer = (nickname ?? "").trim();
}

export function getDeployer(): string {
	return globalDeployer;
}

export function getFooterText(): string {
	const d = getDeployer();
	return d ? `Moebot NEXT (deployed by ${d})` : "Moebot NEXT";
}
