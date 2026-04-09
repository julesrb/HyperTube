import React from "react";

export default function Input({type, placeholder, className }: {type: string, placeholder: string, className?: string}) {
    const isPassword = type === "password";

    return (<div className="flex flex-col w-full">
        <label htmlFor={type} className="text-gray ml-1 font-light">{placeholder}</label>
        {isPassword ?? <a href="#">Forgotten?</a>}
        <input id={type} type={type} placeholder="" className={"rounded-xs p-2 m-0 w-full h-8 bg-white text-black border "  + className}/>
    </div>);
}
