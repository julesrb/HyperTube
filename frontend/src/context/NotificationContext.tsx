"use client";

import React, {createContext, useContext, useState, ReactNode} from "react";


export type tNotificationType = | "success" | "error" | "info" | "warning";

export interface tNotification {
    id: string;
    message: string;
    type: tNotificationType;
}

interface NotificationContextType {
    notifications: tNotification[];
    addNotification: (message: string, type?: tNotificationType) => void;
    removeNotification: (id: string) => void;
}

const NotificationContext = createContext<NotificationContextType | undefined>(undefined);

export const NotificationProvider = ({children,}: { children: ReactNode }) => {
    const [notifications, setNotifications] = useState<tNotification[]>([]);

    const addNotification = (message: string, type: tNotificationType = "info") => {
        const id = crypto.randomUUID();
        const notification: tNotification = {id, message, type};

        setNotifications((prev) => [...prev, notification,]);
        setTimeout(() => {
            removeNotification(id);
        }, 5000);
    };

    const removeNotification = (id: string) => {
        setNotifications((prev) =>
            prev.filter((n) => n.id !== id)
        );
    };

    return (
        <NotificationContext.Provider
            value={{
                notifications,
                addNotification,
                removeNotification,
            }}>{children}
        </NotificationContext.Provider>);
};

export const useNotification = () => {
    const context = useContext(NotificationContext);

    if (!context)
        throw new Error("useNotification must be used inside NotificationProvider");

    return context;
};
