import React from "react";

export default function Input({type, placeholder, className }: {type: string, placeholder: string, className?: string}) {
    const isPassword = type === "password";

    return (<div className="flex flex-col">
        <label htmlFor={type} className="text-gray ml-1 font-light">{placeholder}</label>
        {isPassword ?? <a href="#">Forgotten?</a>}
        <input id={type} type={type} placeholder="" className={"p-2 m-0 w-full h-8 bg-white text-black custom-border "  + className}/>
    </div>);
}
