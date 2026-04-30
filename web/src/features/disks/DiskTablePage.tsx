import type { DiskListItem } from './api';
import { DiskRow } from './DiskRow';

type DiskTablePageProps = {
  disks: DiskListItem[];
};

export function DiskTablePage({ disks }: DiskTablePageProps) {
  return (
    <main className="disk-page">
      <section className="disk-page__panel" aria-labelledby="disk-page-title">
        <div className="disk-page__header">
          <div>
            <p className="disk-page__eyebrow">Disk inventory</p>
            <h1 id="disk-page-title">Disks</h1>
          </div>
          <p className="disk-page__copy">
            Safe metadata stays cached until you request a fresh probe. Wake-capable refresh is explicit so sleeping
            disks are only touched on purpose.
          </p>
        </div>

        <div className="disk-table-wrap">
          <table className="disk-table">
            <thead>
              <tr>
                <th scope="col">Disk</th>
                <th scope="col">State</th>
                <th scope="col">Freshness</th>
                <th scope="col">Actions</th>
              </tr>
            </thead>
            <tbody>
              {disks.map((disk) => (
                <DiskRow key={disk.id} disk={disk} />
              ))}
            </tbody>
          </table>
        </div>
      </section>
    </main>
  );
}
