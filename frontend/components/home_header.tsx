import Image from "next/image";
import Link from "next/link";

type HomeHeaderProps = {
  locationLabel?: string;
};

export function HomeHeader({
  locationLabel = "Stanford, California",
}: HomeHeaderProps = {}) {
  return (
    <tr id="header">
      <td id="h0">
        <Link href="/">
          <Image
            alt="Supostlogo"
            src="/legacy/SUPostLogo.gif"
            width={140}
            height={53}
            priority
          />
        </Link>
      </td>
      <td id="h1">
        <div id="head">
          <form action="/search">
            <input className="searchText" id="q" name="q" type="text" />
            <input type="submit" value="Search" />
          </form>
        </div>
      </td>
      <td id="h_new">
        <div className="header_school">{locationLabel}</div>
      </td>
      <td id="h2">
        <div className="post_button">
          <table>
            <tbody>
              <tr>
                <td>
                  <Link href="/add">post</Link>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </td>
    </tr>
  );
}
