import {NavItem} from "@/types/nav";
import Link from "next/link";
import Image from "next/image";
import styles from "./Navitem.module.css";
import {useState} from "react";

interface Props {
    item: NavItem;
}

export function NavItemComponent({item,}: Props) {
    let iconWidth = 20;
    if (item.icon === "home") iconWidth = 250;
    const Icon = (<Image src={`/icons/${item.icon}.svg`} alt={item.name} width={iconWidth} height={20}/>);

    if ("href" in item) {
        return (<Link className={styles.navitem} href={item.href}>
                {Icon}
                <p>{item.name}</p>
            </Link>);
    }

    if ("hover" in item) {
        const [isOpen, setIsOpen] = useState(false);

        return (
            <div
                className={styles.navitem}
                onMouseEnter={() => (setIsOpen(true))}
                onMouseLeave={() => (setIsOpen(false))}>

                {Icon}
                {isOpen && item.hover()}
            </div>
        );
    }

    return (<button className={styles.navitem + " clean-btn"} onClick={item.action}>
        {Icon}
        <p>{item.name}</p>
    </button>);
}
