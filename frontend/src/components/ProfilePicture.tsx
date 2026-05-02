import {tUser} from "@/types/user";
import React from "react";
import Image from "next/image";

export default function ProfilePicture({user, onClick, size = 0, color, className}: { user: tUser | Partial<tUser>, onClick?: () => void, size?: 0 | 1 | 2, color?: string, className?: string }) {
    const sizes = ["size-10", "size-18 sm:size-24", "size-38 sm:size-45"];
    let children;

    if (user.profile_picture)
        children = <Image className="w-full h-full object-cover" height={200} width={200} src={user.profile_picture} alt="profile picture" />;
    else if (user.firstname && user.lastname){
        const initial = user.firstname[0] + user.lastname[0];

        if (color === undefined)
            color = user.color;

        if (size === 0)
            children = <h6>{initial}</h6>;
        else if (size === 1)
            children = <h3>{initial}</h3>;
        else
            children = <h1 className="text-6xl">{initial}</h1>;
    }

    if (onClick !== undefined)
        className = className + ` hover:bg-${color}-hover`;

    return (<button className={`relative overflow-hidden rounded-full bg-${color} ${sizes[size]} flex items-center justify-center border shrink-0 ` + className} onClick={onClick}>
        {children}
    </button>);
}
