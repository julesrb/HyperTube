"use client";

import NotificationItem from "@/components/ui/Notification/NotificationItem";
import useNotification from "@/contexts/NotificationContext";

export default function NotificationList() {
    const {notifications, removeNotification} = useNotification();

    return (<div className="fixed top-5 right-5 z-60 flex flex-col gap-4 max-w-9/10">
        {notifications.map(notification => (
            <NotificationItem
                key={notification.id}
                notification={notification}
                onClose={removeNotification}
            />
        ))}
    </div>);
}
