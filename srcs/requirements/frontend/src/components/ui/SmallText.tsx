import {ReactNode} from "react";

export default function SmallText({children, className}: {children: ReactNode, className?: string}) {
    return (<p className={"text-xs sm:text-sm text-center italic text-gray " + className}>{children}</p>);
}
