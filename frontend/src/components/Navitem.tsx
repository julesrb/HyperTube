import {NavItem} from "@/types/nav";
import Link from "next/link";
import Image from "next/image";
import styles from "./Navitem.module.css";

interface Props {
    item: NavItem;
}

export function NavItemComponent({item,}: Props) {
    let iconWidth = 20;
    if (item.icon === "home") iconWidth = 250;
    const Icon = (<Image src={`/icons/${item.icon}.svg`} alt={item.name} width={iconWidth} height={20}/>);
    const PName = <p className="hairline">{item.name}</p>
    if ("href" in item) {
        return (<Link className={styles.navitem} href={item.href}>
                {Icon}
                {PName}
            </Link>);
    }

    if ("hover" in item)
        return item.hover(Icon, styles.navitem);

    return (<button className={styles.navitem + " clean-btn"} onClick={item.action}>
        {Icon}
        {PName}
    </button>);
}
