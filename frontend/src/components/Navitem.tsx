import { NavItem } from "@/types/nav";
import Link from "next/link";
import Image from "next/image";


export function NavItemComponent({item,}: { item: NavItem }) {
    const className = "uppercase flex items-center custom-underline";
    const Icon = (<Image className="h-5 w-auto" src={`/icons/${item.icon}.svg`} alt={item.name} width={10} height={10}/>);
    const PName = item.name ? <p className="font-hairline pl-2 text-2xl">{item.name}</p> : null;

    if ("href" in item) {
        return (<Link className={className} href={item.href}>
                {Icon}
                {PName}
            </Link>);
    }

    if ("hover" in item)
        return item.hover(Icon);

    return (<button className={className} onClick={item.action}>
        {Icon}
        {PName}
    </button>);
}
