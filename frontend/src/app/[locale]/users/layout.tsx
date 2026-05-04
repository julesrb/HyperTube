"use client";

import {useAuth} from "@/context/AuthContext";
import {useRouter} from "next/navigation";
import React, {useEffect} from "react";

export default function DashboardLayout({children,}: {children: React.ReactNode;}) {
    const {user, loading} = useAuth();
    const router = useRouter();

    useEffect(() => {
        if (!loading && !user)
            router.push("/");
    }, [user, loading, router]);

    if (loading) return null;
    if (!user) return null;

    return children;
}