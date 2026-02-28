export function HomeFooter() {
  return (
    <tr id="foot">
      <td colSpan={4}>
        <ul id="supost-links">
          <li>
            <a href="http://jobs.supost.com/">post a job</a>
          </li>
          <li>
            <a href="http://housing.supost.com/">post housing</a>
          </li>
          <li>
            <a href="http://cars.supost.com/">post a car</a>
          </li>
          <li>
            <a href="/about">about</a>
          </li>
          <li>
            <a href="/contact">contact</a>
          </li>
          <li>
            <a href="/privacy">privacy</a>
          </li>
          <li>
            <a href="/terms">terms</a>
          </li>
          <li>
            <a href="/help">help</a>
          </li>
        </ul>

        <div id="disclaimer">
          <div>a Greg Wientjes production</div>
          <div>
            SUpost is not sponsored by, endorsed by, or affiliated with Stanford
            University.
          </div>
          <div>SUpost © 2009</div>
        </div>
      </td>
    </tr>
  );
}
