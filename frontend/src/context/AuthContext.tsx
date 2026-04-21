"use client";

import {createContext, useContext, useEffect, useState, ReactNode,} from "react";
import {tUser} from "@/types/user";

type AuthContextType = {
    user: tUser | null;
    login: (user: tUser, token: string) => void;
    logout: () => void;
    loading: boolean;
};

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({children,}: { children: ReactNode; }) {
    const [user, setUser] = useState<tUser | null>(null);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        const token = localStorage.getItem("token");
        const userData = localStorage.getItem("user");

        if (token && userData)
            setUser(JSON.parse(userData));

        setLoading(false);
    }, []);

    const login = (user: tUser, token: string) => {
        localStorage.setItem("token", token);
        localStorage.setItem("user", JSON.stringify(user));
        setUser(user);
    };

    const logout = () => {
        localStorage.removeItem("token");
        localStorage.removeItem("user");
        setUser(null);
    };

    return (<AuthContext.Provider
        value={{user, login, logout, loading,}}>
        {children}
    </AuthContext.Provider>);
}

export function useAuth() {
    const context = useContext(AuthContext);
    if (!context)
        throw new Error("useAuth must be used inside AuthProvider");
    return context;
}