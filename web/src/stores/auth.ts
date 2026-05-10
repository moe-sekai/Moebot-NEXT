import { defineStore } from "pinia";
import {
	getAuthStatus,
	loginAdmin,
	setupAdmin,
	changeAdminPassword,
	setStoredAuthToken,
	getStoredAuthToken,
	type AuthSession,
} from "../api/client";

interface AuthState {
	initialized: boolean | null; // null = 尚未拉取
	username: string;
	nickname: string;
	token: string;
}

// 控制台账号会话状态：缓存 /api/auth/status 结果 + JWT 与昵称。
//
// 设计上只支持单一管理员账号；昵称同时显示在控制台 footer 与所有 Satori
// 渲染卡片底部 footer ("Moebot NEXT (deployed by X)")。
export const useAuthStore = defineStore("auth", {
	state: (): AuthState => ({
		initialized: null,
		username: "",
		nickname: "",
		token: getStoredAuthToken(),
	}),
	getters: {
		isLoggedIn: (s) => Boolean(s.token),
	},
	actions: {
		async refreshStatus() {
			const status = await getAuthStatus();
			this.initialized = status.initialized;
			if (status.nickname) this.nickname = status.nickname;
			if (status.username) this.username = status.username;
			return status;
		},
		applySession(sess: AuthSession) {
			this.token = sess.token;
			this.username = sess.username;
			this.nickname = sess.nickname;
			this.initialized = true;
			setStoredAuthToken(sess.token);
		},
		async login(username: string, password: string) {
			const sess = await loginAdmin(username, password);
			this.applySession(sess);
		},
		async setup(payload: {
			username: string;
			nickname: string;
			password: string;
			password_confirm: string;
		}) {
			const sess = await setupAdmin(payload);
			this.applySession(sess);
		},
		async changePassword(payload: {
			old_password: string;
			new_password: string;
			new_password_confirm: string;
		}) {
			await changeAdminPassword(payload);
		},
		logout() {
			this.token = "";
			setStoredAuthToken("");
		},
	},
});
