"use client";

import React, {useState} from "react";
import {tNotification, useNotification} from "@/context/NotificationContext";
import {CheckIcon, CrossIcon} from "@/components/Icon";

export const NotificationList = () => {
    const { notifications, removeNotification } = useNotification();

    return (<div className="fixed top-5 right-5 z-50 flex flex-col gap-4">
            {notifications.map((notification) => (
                <NotificationItem
                    key={notification.id}
                    notification={notification}
                    onClose={removeNotification}
                />
            ))}
        </div>);
};

const NotificationItem = ({notification, onClose}: { notification: tNotification; onClose: (id: string) => void;}) => {
    const [isHover, setIsHover] = useState(false);
    const bgColors = {success: "green", warning: "yellow", error: "red", info: "purple"};

    return (<div className={`flex justify-between custom-shadow-s border bg-${bgColors[notification.type]}`}>

            <div className="flex items-center min-w-90 max-w-90 border-r p-4">
                {notification.type === "success" && <CheckIcon className="shrink-0 mr-4"/>}
                <p>{notification.message}</p>
            </div>

            <button className="p-4"
                    onClick={() => onClose(notification.id)}
                    onMouseEnter={() => setIsHover(true)}
                    onMouseLeave={() => setIsHover(false)}>
                <CrossIcon color={isHover ? "black-hover" : "black"} />
            </button>
        </div>
    );
};
